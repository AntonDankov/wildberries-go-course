package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
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
	if len(args) != 1 {
		log.Fatal("one file path to text file as argument is required")
	}
	fileName := args[0]

	lines, err := readLinesFromFile(fileName, flags.unique, flags.ignoreBlanks)
	if err != nil {
		log.Fatal(err)
	}

	compareFunc := getComparator(flags)
	if flags.checkSorted {
		isSorted := isSorted(lines, compareFunc)
		if !isSorted {
			fmt.Println("The strings are not sorted from the start")
		}
	}

	sort.SliceStable(lines, func(i, j int) bool {
		return compareFunc(lines[i], lines[j])
	})

	for _, line := range lines {
		fmt.Println(line)
	}
}

func readLinesFromFile(filename string, unique bool, removeTrails bool) ([]string, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	set := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		if removeTrails {
			line = strings.TrimRightFunc(line, unicode.IsSpace)
		}
		if unique {
			if _, exists := set[line]; !exists {
				lines = append(lines, line)
				set[line] = true
			}
		} else {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}
