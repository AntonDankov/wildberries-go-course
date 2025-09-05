package main

import (
	"fmt"
	"sync"
)

func ProceedeGenerate(arr [10]int, output chan int) {
	for _, x := range arr {
		output <- x
	}
	close(output)
}

func ProceedeX2(input chan int, output chan int) {
	for val := range input {
		newVal := val * 2
		output <- newVal
	}
	close(output)
}

func ProceedePrint(input chan int) {
	for val := range input {
		fmt.Printf("Received val: %d\n", val)
	}
}

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	arr := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	conv1Chan := make(chan int, 1)
	conv2Chan := make(chan int, 1)
	go func(arr [10]int, ch chan int) {
		defer waitGroup.Done()
		ProceedeGenerate(arr, conv1Chan)
	}(arr, conv1Chan)

	go func(conv1Chan chan int, convChan2 chan int) {
		defer waitGroup.Done()
		ProceedeX2(conv1Chan, convChan2)
	}(conv1Chan, conv2Chan)

	go func(conv2Chan chan int) {
		defer waitGroup.Done()
		ProceedePrint(conv2Chan)
	}(conv2Chan)

	waitGroup.Wait()
}
