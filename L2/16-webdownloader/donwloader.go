package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

var cssURLRegularExp = regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`)

func putToDownload(donwloadEntity DownloadEntity) {
	wg.Add(1)
	managerDownloadChan <- donwloadEntity
}

func donwload(downloadEntity DownloadEntity, downloadChan chan DownloadEntity) error {
	if strings.HasSuffix(downloadEntity.link, ".css") {
		return downloadCSS(downloadEntity.link, downloadEntity.path, downloadChan)
	}
	return downloadResource(downloadEntity.link, downloadEntity.path)
}

func downloadResource(link string, path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// fmt.Printf("Downloading resource: %s \n", link)
	resp, err := httpClient.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dataToSave, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return saveData(dataToSave, path)
}

func saveData(data []byte, path string) error {
	os.MkdirAll(filepath.Dir(path), 0755)

	// fmt.Printf("Saving as fixed path: %s \n", path)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	out.Write(data)
	out.Close()
	return nil
}

func downloadCSS(link string, path string, downloadChan chan DownloadEntity) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// fmt.Printf("Donwloading css: %s \n", link)
	resp, err := httpClient.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	replacements := make(map[string]string)

	cssBaseURL, _ := url.Parse(link)
	matches := cssURLRegularExp.FindAllSubmatch(data, -1)
	for _, match := range matches {
		relativeLink := string(match[1])
		cleanRelativeLink := relativeLink
		if strings.HasPrefix(relativeLink, "data:") {
			continue
		}
		relativeURL, err := url.Parse(cleanRelativeLink)
		if err != nil {
			return err
		}
		absoluteURL := cssBaseURL.ResolveReference(relativeURL)
		targetPath := convertToLocalPath(absoluteURL.String())

		cssDir := filepath.Dir(path)
		fileRelativePath, _ := filepath.Rel(cssDir, targetPath)
		fileRelativePath = filepath.ToSlash(fileRelativePath)

		replacements[relativeLink] = fileRelativePath
		donwloadEntity := DownloadEntity{
			link: absoluteURL.String(),
			path: targetPath,
		}
		putToDownload(donwloadEntity)
	}
	contentStr := string(data)
	contentStr = cssURLRegularExp.ReplaceAllStringFunc(contentStr, func(match string) string {
		// removing url() and '"
		contentLink := match[4 : len(match)-1]
		contentLink = strings.Trim(contentLink, "'\"")
		if replacementLink, ok := replacements[contentLink]; ok {
			return fmt.Sprintf("url('%s')", replacementLink)
		}
		return match
	})

	data = []byte(contentStr)

	os.MkdirAll(filepath.Dir(path), 0755)

	// fmt.Printf("Saving as fixed path: %s \n", fixedPathStr)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	out.Write(data)
	out.Close()
	return nil
}

func saveHTML(htmlNode *html.Node, path string) error {
	currentFileDir := filepath.Dir(path)
	os.MkdirAll(currentFileDir, 0755)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := html.Render(f, htmlNode); err != nil {
		return err
	}
	return nil
}
