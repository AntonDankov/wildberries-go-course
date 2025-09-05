package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No arguments for amount of workers")
		os.Exit(1)
	}
	amountOfWorkers, err := strconv.Atoi(os.Args[1])
	if err != nil || amountOfWorkers <= 0 {
		fmt.Println("Invalid number of workers")
		os.Exit(1)
	}

	ch := make(chan int, amountOfWorkers)
	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	for i := range amountOfWorkers {
		fmt.Printf("Launching worker: %d\n", i)
		waitGroup.Add(1)
		go func(i int, ch chan int, waitGroup *sync.WaitGroup, ctx context.Context) {
			defer waitGroup.Done()
			for {
				select {
				case value, ok := <-ch:
					if !ok {
						return
					}
					fmt.Printf("%d worker: %d\n", i, value)

				case <-ctx.Done():
					return
				}
			}
		}(i, ch, &waitGroup, ctx)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	running := true

	for running {
		value := rand.Int()
		select {
		case ch <- value:
		case <-signalChan:
			running = false
			cancel()
			close(ch)
		default:
		}
	}
	fmt.Println("Now going to wait")
	waitGroup.Wait()
}
