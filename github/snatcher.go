package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ItemType string

const (
	file      ItemType = "file"
	directory ItemType = "dir"
)

type gitHubItem struct {
	Name       string   `json:"name"`
	Type       ItemType `json:"type"`
	Url        string   `json:"url"`
	ContentUrl string   `json:"download_url"`
}

type gitHubInformation struct {
	userName       string
	repositoryName string
	branch         string
	folderPath     string
}

func (g gitHubInformation) getApiUrl() string {
	var branch string = ""
	if strings.TrimSpace(g.branch) != "" {
		branch = "?ref=" + g.branch
	}
	return fmt.Sprint("https://api.github.com/repos/", g.userName, "/", g.repositoryName, "/contents/", g.folderPath, branch)
}

type Snatcher struct {
	info *gitHubInformation
}

func NewSnatcher(rawUrl string) Snatcher {
	return Snatcher{
		info: extractGitHubInfo(rawUrl),
	}
}

func (s Snatcher) ListContent() {
	fetchAndHandleContent(s.info.getApiUrl(), s.info.folderPath, func(path, contentUrl string) {
		log.Println(green(path), blue(contentUrl))
	})
}

func (s Snatcher) DownloadFolder() {
	fetchAndHandleContent(s.info.getApiUrl(), s.info.folderPath, func(path, contentUrl string) {
		createFolder(path)
		var err = downloadFile(path, contentUrl)
		if err != nil {
			fmt.Printf(red("Invalid URL: %v \n"), err)
		}
		log.Println(blue(path), green("[created]"))
	})
}

func fetchAndHandleContent(url string, path string, handle func(string, string)) {
	items := getGitHubUrlItems(url)

	for _, item := range items {
		var itemPath = filepath.Join(path, item.Name)

		if item.Type == directory {
			fetchAndHandleContent(item.Url, itemPath, handle)

		} else if item.Type == file {
			handle(itemPath, item.ContentUrl)
		}
	}
}

func createFolder(filePath string) string {
	var folderName = filepath.Dir(filePath)
	err := os.MkdirAll(folderName, 0755)
	if err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}
	return folderName
}

func downloadFile(filepath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getGitHubUrlItems(url string) []gitHubItem {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf(red("Error retrieving directory contents: %v"), err)
	}
	defer resp.Body.Close()

	var items []gitHubItem

	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		log.Fatalf(red("Error decoding directory contents: %v"), err)
	}
	return items
}

func extractGitHubInfo(gitHubURL string) *gitHubInformation {
	var parsedURL, err = url.Parse(gitHubURL)
	if err != nil {
		ErrorAndExit(err)
	}

	pattern := `^/([^/]+)/([^/]+)(?:/tree/([^/]+)(?:/(.*))?)?$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(parsedURL.Path)

	if matches == nil {
		ErrorAndExit(fmt.Errorf("incomplete url"))
	}

	return &gitHubInformation{
		userName:       matches[1],
		repositoryName: matches[2],
		branch:         matches[3],
		folderPath:     matches[4],
	}
}

func ErrorAndExit(err error) {
	fmt.Printf(red("Invalid URL: %v \n"), err)
	fmt.Print(
		"Expected URLs:\n",
		green("https://github.com/<username>/<repo>/"),
		yellow(" => Download all the repository.\n"),
		green("https://github.com/<username>/<repo>/tree/master/<folder/path>"),
		yellow(" => Download a folder.\n"))
	os.Exit(1)
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[34m" + s + "\033[0m"
}
