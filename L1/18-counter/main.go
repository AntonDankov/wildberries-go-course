package main

import (
	"sync/atomic"
)

type AtomicCounter struct {
	counter int32
}

func (ac *AtomicCounter) IncrementCounter() int32 {
	return atomic.AddInt32(&ac.counter, 1)
}
