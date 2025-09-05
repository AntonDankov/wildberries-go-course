package main

import "fmt"

func getIntersection[T comparable](slice1 []T, slice2 []T) []T {
	mp1 := covertSliceToMap(slice1)
	mp2 := covertSliceToMap(slice2)

	resultSlice := []T{}
	for key := range mp1 {
		_, exists := mp2[key]
		if exists {
			resultSlice = append(resultSlice, key)
		}
	}

	return resultSlice
}

func printSlice[T any](slice []T) {
	for _, value := range slice {
		fmt.Printf("%v, ", value)
	}
}

func covertSliceToMap[T comparable](slice []T) map[T]bool {
	mp := make(map[T]bool)
	for _, value := range slice {
		mp[value] = true
	}
	return mp
}

func main() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{2, 3, 4}

	resSlice := getIntersection(slice1, slice2)
	printSlice(resSlice)
}
