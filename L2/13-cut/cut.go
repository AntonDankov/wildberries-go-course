package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Range struct {
	start int
	end   int
}

// Flags struct to operate with program
type Flags struct {
	fieldRanges []Range
	delimeter   string
	separated   bool
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags() Flags {
	var flags Flags
	var rangeString string

	flag.StringVar(&flags.delimeter, "d", "\t", "Delimeter, default '\t'")
	flag.StringVar(&rangeString, "f", "", "Fields number what should be printed")

	flag.BoolVar(&flags.separated, "s", false, "Print rows what has delimeter")
	flag.Parse()

	if rangeString == "" {
		log.Fatal("no fields specified (-f flag)")
	}
	parts := strings.Split(strings.Trim(rangeString, " "), ",")

	if len(parts) == 0 {
		log.Fatal("no fields specified (-f flag)")
	} else {
		for _, part := range parts {
			splitPart := strings.Split(part, "-")
			if len(splitPart) == 0 {
				log.Fatal("not specified field correctly")
			} else if len(splitPart) == 1 {
				value, err := strconv.Atoi(splitPart[0])
				if err != nil {
					log.Fatal(part)
				}
				value--
				fieldRange := Range{
					start: value,
					end:   value,
				}
				flags.fieldRanges = append(flags.fieldRanges, fieldRange)
			} else if len(splitPart) > 2 {
				log.Fatal("not correct range")
			} else {
				valueStart, err := strconv.Atoi(splitPart[0])
				if err != nil {
					log.Fatal(err)
				}
				valueStart--
				valueEnd, err := strconv.Atoi(splitPart[1])
				if err != nil {
					log.Fatal(err)
				}
				valueEnd--
				fieldRange := Range{
					start: valueStart,
					end:   valueEnd,
				}

				flags.fieldRanges = append(flags.fieldRanges, fieldRange)
			}
		}
	}
	return flags
}

func processLine(flags Flags, line string) {
	var sb strings.Builder
	fields := strings.Split(line, flags.delimeter)
	if len(fields) <= 1 && flags.separated {
		return
	}
	for _, fieldRange := range flags.fieldRanges {
		for i := fieldRange.start; i < fieldRange.end+1; i++ {
			if i >= len(fields) {
				break
			}
			sb.WriteString(fields[i] + flags.delimeter)
		}
	}
	sb.WriteString("\n")
	fmt.Print(sb.String())
}
