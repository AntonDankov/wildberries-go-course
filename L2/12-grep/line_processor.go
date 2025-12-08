package main

import (
	"bufio"
	"flag"
	"os"
	"regexp"
	"strings"
)

// Flags struct to operate with programm
type Flags struct {
	after           int
	before          int
	countOnly       bool
	caseIgnore      bool
	invert          bool
	fixedString     bool
	printNumberLine bool
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags() Flags {
	var flags Flags
	flag.IntVar(&flags.after, "A", 0, "print N lines after match")
	flag.IntVar(&flags.before, "B", 0, "print N lines before match")
	var around int
	flag.IntVar(&around, "C", 0, "print N lines before and after match")
	if around != 0 {
		flags.after = around
		flags.before = around
	}

	flag.BoolVar(&flags.countOnly, "c", false, "only print count of matching lines")
	flag.BoolVar(&flags.caseIgnore, "i", false, "ignore case")
	flag.BoolVar(&flags.invert, "v", false, "invert match")
	flag.BoolVar(&flags.fixedString, "F", false, "fixed string match instead of regexp")
	flag.BoolVar(&flags.printNumberLine, "n", false, "print line number")
	flag.Parse()
	return flags
}

func getMatcher(flags Flags, pattern string) func(string) bool {
	var matchFunc func(string, string, bool) bool
	if flags.fixedString {
		matchFunc = fixedStringMatch
		if flags.caseIgnore {
			pattern = strings.ToLower(pattern)
		}
	} else {
		matchFunc = patternStringMatch
		if flags.caseIgnore {
			caseIgnorePattern := "(?i)"
			pattern = caseIgnorePattern + pattern
		}
	}
	return func(line string) bool {
		result := matchFunc(line, pattern, flags.caseIgnore)

		if flags.invert {
			return !result
		}
		return result
	}
}

func fixedStringMatch(line string, pattern string, caseIgnore bool) bool {
	if caseIgnore {
		line = strings.ToLower(line)
	}

	return line == pattern
}

func patternStringMatch(line string, pattern string, caseIgnore bool) bool {
	exp, err := regexp.Compile(pattern)
	if err != nil {
		panic("Invalid regex")
	}

	return exp.MatchString(line)
}

func collectLines(input *os.File, matcher func(string) bool) ([]string, []int) {
	scanner := bufio.NewScanner(input)
	var lines []string
	var matchedIds []int
	counter := 0

	for scanner.Scan() {
		line := scanner.Text()
		if matcher(line) {
			matchedIds = append(matchedIds, counter)
		}
		lines = append(lines, line)
		counter++
	}

	return lines, matchedIds
}
