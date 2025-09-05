package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Контектс более предпочтительный способ для завершения работы горутин
	// является стандартом для завершение, когда канал может использоваться для абсолютно разных вещей
	// есть специальные методы такие как timeout и deadline
	// Каналы используются в специальных случаях (возможно когда стот передать не только сигнал о завершении, но еще и какое-то значение, которое нужно использовать при завершении)
	ctx, cancel := context.WithCancel(context.Background())

	var waitGroup sync.WaitGroup
	for i := range 3 {
		waitGroup.Add(1)
		go func(i int, ctx context.Context, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					fmt.Printf("Worker %d\n", i)
				}
			}
		}(i, ctx, &waitGroup)
	}

	running := true
	for running {
		select {
		case <-sigChan:
			cancel()
			running = false
		default:
			fmt.Printf("Main goroutine\n")
		}
	}
	waitGroup.Wait()
}
