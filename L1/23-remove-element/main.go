package main

import "fmt"

func removeElementAtIndex[T any](slice []T, i int) []T {
	N := len(slice)
	if i < 0 || i >= N {
		return slice
	}
	copy(slice[i:], slice[i+1:])
	return slice[:N-1]
}

func main() {
	all := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("all: ", all)
	removeIndex := removeElementAtIndex(all, 5)
	fmt.Println("removeIndex: ", removeIndex)
}
