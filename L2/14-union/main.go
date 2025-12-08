package main

import (
	"fmt"
	"time"
)

// Non-blocking union of channels
// func or(channels ...<-chan interface{}) <-chan interface{} {
// 	for {
// 		for _, channel := range channels {
// 			select {
// 			case <-channel:
// 				return channel
// 			default:
// 				continue
// 			}
// 		}
// 	}
// }

func or(channels ...<-chan interface{}) <-chan interface{} {
	N := len(channels)
	switch N {
	case 0:
		return nil
	case 1:
		return channels[0]
	case 2:
		done := make(chan interface{})
		go func() {
			defer close(done)
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		}()
		return done

	default:
		divide := N / 2
		return or(or(channels[:divide]...), or(channels[divide:]...))

	}
}

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}
