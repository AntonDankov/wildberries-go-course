package main

import (
	"fmt"
)

func groupValues(mp map[int][]float64, slice []float64) {
	for _, val := range slice {
		key := int(val) - (int(val) % 10)

		mp[key] = append(mp[key], val)
	}
}

func printMap(mp map[int][]float64) {
	for key, values := range mp {
		fmt.Printf("%d:{", key)
		size := len(values)
		for i, val := range values {
			fmt.Printf("%.1f", val)
			if i < size-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Printf("}\n")
	}
}

func main() {
	arr := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5, -29.9, 0, 1}
	mp := make(map[int][]float64)
	groupValues(mp, arr)
	printMap(mp)
}
