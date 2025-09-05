package main

import "fmt"

func flipWords(s []rune) {
	flipString(s)
	prevIndex := 0
	curIndex := 0
	N := len(s)
	for curIndex <= N {
		if curIndex == N || s[curIndex] == ' ' {
			r := curIndex - 1
			for prevIndex < r {
				buf := s[prevIndex]
				s[prevIndex] = s[r]
				s[r] = buf
				prevIndex++
				r--
			}
			curIndex++
			prevIndex = curIndex
		} else {
			curIndex++
		}
	}
}

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
	s := []rune("snow dog sun")
	flipWords(s)
	fmt.Printf("%s\n", string(s))
	s = []rune("  ")
	flipWords(s)
	fmt.Printf("%s\n", string(s))

	s = []rune(" sun  snow ")
	flipWords(s)
	fmt.Printf("%s\n", string(s))
}
