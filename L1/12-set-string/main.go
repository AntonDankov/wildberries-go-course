package main

import "fmt"

func printSlice[T any](slice []T) {
	for _, value := range slice {
		fmt.Printf("%v, ", value)
	}
}

func convertSliceToSet[T comparable](slice []T) []T {
	mp := convertSliceToMap(slice)
	resSlice := []T{}
	for key := range mp {
		resSlice = append(resSlice, key)
	}
	return resSlice
}

func convertSliceToMap[T comparable](slice []T) map[T]bool {
	mp := make(map[T]bool)
	for _, value := range slice {
		mp[value] = true
	}
	return mp
}

func main() {
	slice := []string{"cat", "cat", "dog", "cat", "tree"}

	resSlice := convertSliceToSet(slice)
	printSlice(resSlice)
}
