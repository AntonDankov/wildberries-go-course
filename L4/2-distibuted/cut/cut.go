package cut

import (
	"flag"
	"log"
	"strconv"
	"strings"
)

type Range struct {
	Start int
	End   int
}

// Flags struct to operate with program
type Flags struct {
	FieldRanges []Range
	Delimeter   string
	Separated   bool
	InputFile   string
	OutputFile  string
}

// GetFlags parsing flag passed by user and set Flags struct
func GetFlags(input string) Flags {
	var flags Flags
	var rangeString string

	args := strings.Fields(input)

	flags.InputFile = args[0]
	args = args[1:]

	flagSet := flag.NewFlagSet("cut", flag.ContinueOnError)

	flagSet.StringVar(&flags.Delimeter, "d", "\t", "Delimeter, default '\\t'")
	flagSet.StringVar(&rangeString, "f", "", "Fields number what should be printed")

	flagSet.BoolVar(&flags.Separated, "s", false, "Print rows what has delimeter")
	flagSet.StringVar(&flags.OutputFile, "o", "", "Path to output file, if not specified, output will be in the console")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

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
					Start: value,
					End:   value,
				}
				flags.FieldRanges = append(flags.FieldRanges, fieldRange)
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
					Start: valueStart,
					End:   valueEnd,
				}

				flags.FieldRanges = append(flags.FieldRanges, fieldRange)
			}
		}
	}
	return flags
}

func ProcessLine(flags Flags, line string, stringBuilder *strings.Builder) {
	fields := strings.Split(line, flags.Delimeter)
	if len(fields) <= 1 && flags.Separated {
		return
	}
	for _, fieldRange := range flags.FieldRanges {
		for i := fieldRange.Start; i < fieldRange.End+1; i++ {
			if i >= len(fields) {
				break
			}
			stringBuilder.WriteString(fields[i] + flags.Delimeter)
		}
	}
	stringBuilder.WriteString("\n")
}
