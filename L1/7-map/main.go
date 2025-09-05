package main

import (
	"fmt"
	"sync"
)

type SyncMap[T any] struct {
	hashmap map[string]T
	mutex   sync.RWMutex
}

func NewSyncMap[T any]() *SyncMap[T] {
	return &SyncMap[T]{
		hashmap: make(map[string]T),
	}
}

func (syncMap *SyncMap[T]) Add(key string, val T) {
	syncMap.mutex.Lock()
	defer syncMap.mutex.Unlock()
	syncMap.hashmap[key] = val
}

func main() {
	syncMap := NewSyncMap[int]()

	syncMap.Add("test1", 1)
	syncMap.Add("test2", 2)
	syncMap.Add("test3", 3)

	var waitGroup sync.WaitGroup

	for i := range 5 {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			for j := range 5 {
				key := fmt.Sprintf("%d:%d", j, i)
				syncMap.Add(key, i*j)
			}
		}(i)
	}

	waitGroup.Wait()
}
