package main

import (
	"fmt"
	"sort"
	"strings"
)

func sortString(str string) string {
	s := strings.Split(str, "")
	sort.Strings(s)
	return strings.Join(s, "")
}

func findAnagrams(words []string) map[string][]string {
	mp := make(map[string][]string)
	set := make(map[string]bool)
	firstAppear := make(map[string]string)

	for _, word := range words {
		lowerCaseWord := strings.ToLower(word)
		sortedWord := sortString(lowerCaseWord)
		if _, exists := set[lowerCaseWord]; !exists {
			key, exists := firstAppear[sortedWord]
			if !exists {
				key = lowerCaseWord
				firstAppear[sortedWord] = lowerCaseWord
			}
			mp[key] = append(mp[key], lowerCaseWord)
			set[lowerCaseWord] = true
		}

	}

	anagrams := make(map[string][]string)
	for anagram, list := range mp {
		if len(list) < 2 {
			continue
		}

		sort.Strings(list)
		anagrams[anagram] = list
	}

	return anagrams
}

func main() {
	input := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	result := findAnagrams(input)

	for key, val := range result {
		fmt.Printf("%q: %v\n", key, val)
	}
}
