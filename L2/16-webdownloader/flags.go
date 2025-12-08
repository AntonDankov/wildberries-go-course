package main

import (
	"flag"
	"log"
)

// Flags struct to operate with program
type Flags struct {
	link          string
	depth         int
	threadsAmount int
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags() Flags {
	var flags Flags

	flag.StringVar(&flags.link, "u", "", "Start URL for crawling")
	flag.IntVar(&flags.depth, "d", 1, "Depth to travers (min 1)")
	flag.IntVar(&flags.threadsAmount, "N", 1, "amount of threads to crawl and download")

	flag.Parse()

	if flags.link == "" {
		log.Fatal("no fields specified (-u flag)")
	}
	if flags.depth < 1 {
		log.Fatal("Depth flag value should be min 1")
	}
	if flags.threadsAmount < 1 {
		log.Fatal("Amoutn of threads flag value should be min 1")
	}

	return flags
}
