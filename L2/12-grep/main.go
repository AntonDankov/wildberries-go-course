package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func preprocessArgs() {
	orig := os.Args
	var expanded []string
	expanded = append(expanded, orig[0])
	for _, arg := range orig[1:] {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			for _, c := range arg[1:] {
				expanded = append(expanded, "-"+string(c))
			}
		} else {
			expanded = append(expanded, arg)
		}
	}
	os.Args = expanded
}

func main() {
	preprocessArgs()
	flags := GetFlags()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("patter was not provided")
	}

	pattern := args[0]

	var input *os.File
	input = os.Stdin

	if len(args) == 2 {
		fileInput, err := os.Open(args[1])
		if err != nil {
			log.Fatal(err)
		}
		input = fileInput

	}

	matcher := getMatcher(flags, pattern)

	lines, ids := collectLines(input, matcher)

	if flags.countOnly {
		fmt.Println(len(ids))
		return
	}
	prevStart := -1
	for _, id := range ids {
		start := max(prevStart, max(0, id-flags.before))
		end := min(id+flags.after+1, len(lines))
		for i := start; i < end; i++ {
			if flags.printNumberLine {
				fmt.Printf("%v : %s\n", i+1, lines[i])
			} else {
				fmt.Println(lines[i])
			}
		}
		prevStart = end
	}
}
