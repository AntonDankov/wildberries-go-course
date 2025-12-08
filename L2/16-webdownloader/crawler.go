package main

import (
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func putToProcessLink(linkEntity LinkEntity) {
	wg.Add(1)
	// fmt.Printf("Putting link %s\n", linkEntity.url)
	managerLinkChan <- linkEntity
}

func processLink(linkEntity LinkEntity, linkChan chan LinkEntity, downloadChan chan DownloadEntity) error {
	if linkEntity.level >= MAX_LEVEL {
		return nil
	}
	resp, err := httpClient.Get(linkEntity.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	u, err := url.Parse(linkEntity.url)
	if err != nil {
		return err
	}
	localPath := convertToLocalPath(linkEntity.url)
	currentFileDir := filepath.Dir(localPath)

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		donwloadEntity := DownloadEntity{
			link: linkEntity.url,
			path: localPath,
		}
		putToDownload(donwloadEntity)
	} else {

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return err
		}

		processHTTPPage(doc, u, currentFileDir, linkEntity.level, linkChan, downloadChan)

		return saveHTML(doc, localPath)
	}
	return nil
}

func processHTTPPage(rootNode *html.Node, baseURL *url.URL, currentFileDir string, level int, linkChan chan LinkEntity, downloadChan chan DownloadEntity) {
	var queue []*html.Node
	queue = append(queue, rootNode)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.Type == html.ElementNode {
			var key string
			switch node.Data {
			case "a", "link":
				key = "href"
			case "img", "script":
				key = "src"
			}

			if key != "" {
				for i, attr := range node.Attr {
					if attr.Key == key {
						linkURL, err := url.Parse(attr.Val)
						if err != nil {
							continue
						}
						absoluteURL := baseURL.ResolveReference(linkURL)

						targetLocalPath := convertToLocalPath(absoluteURL.String())

						if node.Data != "a" {
							donwloadEntity := DownloadEntity{
								link: absoluteURL.String(),
								path: targetLocalPath,
							}
							putToDownload(donwloadEntity)
						} else if absoluteURL.Host == baseURL.Host {
							linkEntity := LinkEntity{
								url:   absoluteURL.String(),
								level: level + 1,
							}
							putToProcessLink(linkEntity)
						}

						relativePath, _ := filepath.Rel(currentFileDir, targetLocalPath)
						relativePath = filepath.ToSlash(relativePath)
						node.Attr[i].Val = relativePath
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			queue = append(queue, child)
		}
	}
}
