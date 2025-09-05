package main

import "fmt"

func quicksort(slice []int) {
	N := len(slice)

	if N <= 1 {
		return
	}
	pivot := slice[(len(slice)-1)/2]
	leftIndex := 0
	rightIndex := N - 1
	for leftIndex <= rightIndex {
		for slice[leftIndex] < pivot {
			leftIndex++
		}
		for slice[rightIndex] > pivot {
			rightIndex--
		}
		if leftIndex <= rightIndex {
			slice[leftIndex], slice[rightIndex] = slice[rightIndex], slice[leftIndex]
			leftIndex++
			rightIndex--
		}

	}
	quicksort(slice[:leftIndex])
	quicksort(slice[leftIndex:])
}

func main() {
	arr := []int{5, 1, 1, 2, 0, 0}
	quicksort(arr)
	fmt.Printf("%v\n", arr)
}
