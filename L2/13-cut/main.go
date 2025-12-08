package main

import (
	"bufio"
	"os"
)

func main() {
	flags := GetFlags()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		processLine(flags, line)
	}
}
