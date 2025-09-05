package main

import (
	"fmt"
	"sync"
)

func calculate(index int, arr *[5]int) {
	number := (*arr)[index]
	res := number * number
	fmt.Printf("%d : %d\n", index, res)
}

func main() {
	numbers := [5]int{2, 4, 6, 8, 10}

	var wg sync.WaitGroup
	for i := range numbers {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			calculate(index, &numbers)
		}(i)

	}
	wg.Wait()
}
