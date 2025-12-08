package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func unpack(str string) (string, error) {
	if str == "" {
		return str, nil
	}

	runes := []rune(str)
	var stringBuilder strings.Builder
	var bufRune rune
	lastRuneFilled := false

	counter := 0

	N := len(runes)

	for i := 0; i <= N; i++ {
		if i == N {
			if counter > 0 && !lastRuneFilled {
				return "", errors.New("no character was provided to unpack")
			}
			amount := max(1, counter)
			for j := 0; j < amount; j++ {
				stringBuilder.WriteRune(bufRune)
			}
			break
		}
		if unicode.IsDigit(runes[i]) {
			digit := int(runes[i]) - '0'

			counter = counter*10 + digit
		} else {
			if lastRuneFilled {
				amount := max(1, counter)
				for j := 0; j < amount; j++ {
					stringBuilder.WriteRune(bufRune)
				}
				counter = 0
			}
			bufRune = runes[i]
			lastRuneFilled = true
		}
	}

	return stringBuilder.String(), nil
}

func main() {
	testval := "a4bc2d5e"
	res, _ := unpack(testval)
	fmt.Printf("%s: %s\n", testval, res)
}
