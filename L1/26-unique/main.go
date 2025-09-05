package main

import (
	"fmt"
	"unicode"
)

func checkUnique(str string) bool {
	mp := make(map[rune]bool)
	for _, char := range str {
		lowChar := unicode.ToLower(char)
		_, exist := mp[lowChar]
		if exist {
			return false
		} else {
			mp[lowChar] = true
		}

	}

	return true
}

func main() {
	str := "aBcdE"
	res := checkUnique(str)
	fmt.Println(res)
	str = "aABcdE"
	res = checkUnique(str)
	fmt.Println(res)
	str = "aaBcdE"
	res = checkUnique(str)
	fmt.Println(res)
	str = ""
	res = checkUnique(str)
	fmt.Println(res)
}
