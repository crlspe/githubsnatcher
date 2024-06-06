package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/crlspe/githubsnatcher/github"
)

func main() {

	var download = flag.Bool("download", true, "Download GitHUb Repository sub-folder.")
	var dryrun = flag.Bool("dry-run", false, "List files and contentUrls to be download.")

	flag.Parse()

	var args = flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: githubsnatcher [--download] [--dry-run] <url>")
		os.Exit(1)
	}

	var rawUrl = args[0]

	var snatcher = github.NewSnatcher(rawUrl)
	switch {
	case *dryrun:
		{
			snatcher.ListContent()
		}
	case *download:
		{
			snatcher.DownloadFolder()
		}
	}
}
