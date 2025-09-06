package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func conditionalExit() {
	counter := 0

	for {
		counter++
		if counter == 10000 {
			fmt.Println("Goroutine with conditional exit EXITS...")
			return
		}

	}
}

func channelExit(ch chan struct{}) {
	for {
		select {
		case <-ch:
			fmt.Println("Goroutine with channelExit EXITS")
			return
		default:
			fmt.Println("Goroutine with channelExit still working...")
			time.Sleep(time.Second)
		}
	}
}

func contextExit(context context.Context) {
	for {
		select {
		case <-context.Done():
			fmt.Println("Goroutine with contexExit EXITS")
			return
		default:
			fmt.Println("Goroutine with contextExit still working...")
			time.Sleep(time.Second)
		}
	}
}

func runtimeExit() {
	time.Sleep(time.Second)
	fmt.Println("Goroutine with runtime exit EXITS...")
	runtime.Goexit()
	fmt.Println("Goroutine with runtime didnt exit properly...")
}

func main() {
	var waitgroup sync.WaitGroup

	waitgroup.Go(func() {
		conditionalExit()
	})
	ch := make(chan struct{})
	waitgroup.Go(func() {
		channelExit(ch)
	})
	ctx, cancel := context.WithCancel(context.Background())

	waitgroup.Go(func() {
		contextExit(ctx)
	})
	waitgroup.Go(func() {
		runtimeExit()
	})
	close(ch)
	cancel()

	fmt.Println("Waiting for goroutines to finish")
	waitgroup.Wait()
}
