package main

import (
	"flag"
	"log"
)

// Flags struct to operate with program
type Flags struct {
	host    string
	port    string
	timeout int
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags() Flags {
	var flags Flags

	flag.StringVar(&flags.host, "host", "", "Host value")
	flag.StringVar(&flags.port, "port", "", "Host port")

	flag.IntVar(&flags.timeout, "timeout", 10, "Timeout in seconds")

	flag.Parse()

	if flags.host == "" {
		log.Fatal("no host specified")
	}
	if flags.port == "" {
		log.Fatal("no port specified")
	}
	if flags.timeout < 1 {
		log.Fatal("Timeout cant be less than 1")
	}

	return flags
}
