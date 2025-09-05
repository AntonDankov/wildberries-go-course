package main

import "fmt"

func binarySearch(arr []int, x int) int {
	left := 0
	right := len(arr) - 1
	for left < right {
		mid := int((left + right) / 2)
		if x > arr[mid] {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if arr[left] == x {
		return left
	} else {
		return -1
	}
}

func main() {
	arr := []int{-5, 0, 3, 4, 6, 10, 20, 23, 24, 30}

	ans := binarySearch(arr, -6)
	fmt.Printf("index: %d \n", ans)
	ans = binarySearch(arr, -5)
	fmt.Printf("index: %d \n", ans)
	ans = binarySearch(arr, 3)
	fmt.Printf("index: %d \n", ans)
	ans = binarySearch(arr, 30)
	fmt.Printf("index: %d \n", ans)
	ans = binarySearch(arr, 31)
	fmt.Printf("index: %d \n", ans)
}
