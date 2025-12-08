package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Flags struct to operate with programm
type Flags struct {
	column                int
	sortByNumber          bool
	reverse               bool
	unique                bool
	sortByMonth           bool
	ignoreBlanks          bool
	checkSorted           bool
	sortByNumberAndSuffix bool
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags() Flags {
	var flags Flags
	flag.IntVar(&flags.column, "k", 1, "sort by column N (1-based, tab separated)")
	flag.BoolVar(&flags.sortByNumber, "n", false, "sort numerically")
	flag.BoolVar(&flags.reverse, "r", false, "reverse order")
	flag.BoolVar(&flags.unique, "u", false, "unique lines only")
	flag.BoolVar(&flags.sortByMonth, "M", false, "sort by month name")
	flag.BoolVar(&flags.ignoreBlanks, "b", false, "ignore trailing blanks")
	flag.BoolVar(&flags.checkSorted, "c", false, "check if sorted")
	flag.BoolVar(&flags.sortByNumberAndSuffix, "h", false, "human-readable number sort with suffixes")
	flag.Parse()
	return flags
}

var months = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var suffixes = map[rune]int{
	'k': 1 << 10, 'm': 1 << 20, 'g': 1 << 30,
}

func getComparator(flags Flags) func(a string, b string) bool {
	var compareFunc func(string, string) bool
	switch {
	case flags.sortByMonth:
		compareFunc = compareByMonth
	case flags.sortByNumber:
		compareFunc = compareByNumber
	case flags.sortByNumberAndSuffix:
		compareFunc = compareByNumberWithSuffix
	default:
		compareFunc = compareDefault
	}
	return func(a string, b string) bool {
		columnA := extractKey(a, flags.column)
		columnB := extractKey(b, flags.column)

		compareResult := compareFunc(columnA, columnB)

		if flags.reverse {
			return !compareResult
		}
		return compareResult
	}
}

func extractKey(line string, key int) string {
	columns := strings.Split(line, "\t")
	if key < 1 || key > len(columns) {
		return line
	}
	column := columns[key-1]
	return column
}

func compareByMonth(a string, b string) bool {
	monthA := months[strings.ToLower(a[:3])]
	monthB := months[strings.ToLower(b[:3])]
	if monthA == monthB {
		return a < b
	}
	return monthA < monthB
}

func compareByNumber(a string, b string) bool {
	numberA, err := strconv.Atoi(a)
	if err != nil {
		panic(err)
	}
	numberB, err := strconv.Atoi(b)
	if err != nil {
		panic(err)
	}
	return numberA < numberB
}

func compareByNumberWithSuffix(a string, b string) bool {
	sufA, err := parseSuffix(a)
	if err != nil {
		panic(err)
	}
	sufB, err := parseSuffix(b)
	if err != nil {
		panic(err)
	}
	return sufA < sufB
}

func compareDefault(a string, b string) bool {
	return a < b
}

func parseSuffix(str string) (int, error) {
	N := len(str)
	if N == 0 {
		return 0, errors.New("empty string, no suffix")
	}
	suffix := unicode.ToLower(rune(str[N-1]))
	sufNumber, exists := suffixes[suffix]
	if !exists {
		return 0, fmt.Errorf("bad suffix '%v' provided", suffix)
	}
	number, err := strconv.Atoi(str[:N-1])
	if err != nil {
		return 0, err
	}
	number = number * sufNumber
	return number, nil
}

func isSorted(lines []string, compare func(a string, b string) bool) bool {
	N := len(lines)
	for i := 1; i < N; i++ {
		if !compare(lines[i-1], lines[i]) {
			return false
		}
	}
	return true
}
