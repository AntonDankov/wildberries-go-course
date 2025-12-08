package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	Port string
}

func GetFlags() Flags {
	var flags Flags

	flag.StringVar(&flags.Port, "port", "", "Host port")

	flag.Parse()

	if flags.Port == "" {
		fmt.Println("No port was provided with -port flag, will be used default 8080 port")
		flags.Port = "8080"
	}

	return flags
}
