package main

import "fmt"

func flipString(s []rune) {
	l := 0
	r := len(s) - 1
	for l < r {
		leftRune := s[l]
		s[l] = s[r]
		s[r] = leftRune
		l++
		r--
	}
}

func main() {
	s := []rune("главрыба")
	flipString(s)
	fmt.Printf("%s\n", string(s))
}
