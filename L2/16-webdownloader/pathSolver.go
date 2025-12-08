package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

// DefaultPathToSave folder where the file going to be saved
const DefaultPathToSave = "scraped"

func convertToLocalPath(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		panic(err)
	}
	u.RawQuery = ""

	host := strings.ReplaceAll(u.Host, ":", "_")
	path := u.Path
	if index := strings.Index(path, "?"); index != -1 {
		path = path[:index]
	}

	if path == "" || strings.HasSuffix(path, "/") {
		path += "index.html"
	} else if filepath.Ext(path) == "" {
		path += ".html"
	}
	localPath := filepath.Join(DefaultPathToSave, host, path)

	return filepath.Clean(localPath)
}
