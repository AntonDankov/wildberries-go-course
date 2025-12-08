package main

import (
	"sync"
)

// DownloadEntity contains link and path to save file
type DownloadEntity struct {
	link string
	path string
}

// LinkEntity contains url and depth level
type LinkEntity struct {
	url   string
	level int
}

// MaxLevel to traverse
var MaxLevel int

var (
	linkChan            = make(chan LinkEntity, 1)
	downloadChan        = make(chan DownloadEntity, 1)
	managerLinkChan     = make(chan LinkEntity, 1)
	managerDownloadChan = make(chan DownloadEntity, 1)
	wg                  sync.WaitGroup
)

func manage() {
	visited := make(map[string]bool)
	var linkQueue []LinkEntity
	var downloadQueue []DownloadEntity
	for {
		var linkOutput chan LinkEntity
		var nextLink LinkEntity
		var downloadOutput chan DownloadEntity
		var nextDownload DownloadEntity
		if len(linkQueue) > 0 {
			linkOutput = linkChan
			nextLink = linkQueue[0]
		}
		if len(downloadQueue) > 0 {
			downloadOutput = downloadChan
			nextDownload = downloadQueue[0]
		}
		select {
		case linkEntity := <-managerLinkChan:
			_, ok := visited[linkEntity.url]
			if ok {
				// fmt.Printf("already processed this link %s\n", newLink.url)
				wg.Done()
			} else if linkEntity.level >= MaxLevel {
				wg.Done()
			} else {
				visited[linkEntity.url] = true
				linkQueue = append(linkQueue, linkEntity)
			}
		case downloadEntity := <-managerDownloadChan:
			_, ok := visited[downloadEntity.link]
			if ok {
				// fmt.Printf("already downloaded this %s\n", downloadEntity.link)
				wg.Done()
			} else {
				visited[downloadEntity.link] = true
				downloadQueue = append(downloadQueue, downloadEntity)
			}
		case linkOutput <- nextLink:
			linkQueue = linkQueue[1:]
		case downloadOutput <- nextDownload:
			downloadQueue = downloadQueue[1:]
		}

	}
}
