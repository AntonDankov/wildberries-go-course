package main

import (
	"fmt"
	"os"

	"github.com/beevik/ntp"
)

func main() {
	time, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NTP time request error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("NTP time: %v", time)
}
