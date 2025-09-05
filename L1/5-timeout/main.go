package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	timeoutSeconds, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Invalid timeout value: %v\n", err)
		os.Exit(1)
	}
	timer := time.NewTimer(time.Duration(timeoutSeconds) * time.Second)

	ch := make(chan int, 1)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func(ch chan int, ctx context.Context) {
		defer waitGroup.Done()
		for {
			select {
			case val := <-ch:
				{
					fmt.Println(val)
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}(ch, ctx)
	running := true
	for running {
		select {
		case <-timer.C:
			running = false
		default:
			ch <- rand.Int()
		}
	}
	cancel()
	waitGroup.Wait()
}
