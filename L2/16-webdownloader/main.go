package main

import (
	"fmt"
)

func main() {
	// link to test
	// link := "https://books.toscrape.com/catalogue/a-light-in-the-attic_1000/index.html"
	flags := GetFlags()
	MAX_LEVEL = flags.depth

	for range flags.threadsAmount {
		go func() {
			for linkEntity := range linkChan {

				fmt.Printf("Getting link %s\n", linkEntity.url)
				err := processLink(linkEntity, linkChan, downloadChan)
				if err != nil {
					fmt.Printf("ERROR while crawling: %v\n", err)
				}
				wg.Done()
			}
		}()
		go func() {
			for downloadEntity := range downloadChan {
				err := donwload(downloadEntity, downloadChan)
				if err != nil {
					fmt.Printf("ERROR while downloading: %v\n", err)
				}
				wg.Done()
			}
		}()
	}

	linkEntity := LinkEntity{
		url:   flags.link,
		level: 0,
	}
	go func() {
		manage()
	}()
	putToProcessLink(linkEntity)
	wg.Wait()
	close(linkChan)
	close(downloadChan)
}
