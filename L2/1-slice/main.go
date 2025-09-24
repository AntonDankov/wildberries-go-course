package main

import "fmt"

func main() {
	a := [5]int{76, 77, 78, 79, 80}
	// Берется слайся от исходного массива где
	// 1-lower bound index (включительно) и 4 - upper bound index (невключительно)
	// [1;4)
	var b []int = a[1:4]

	fmt.Println(b)
}
