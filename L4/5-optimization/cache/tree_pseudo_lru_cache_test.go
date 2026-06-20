package cache

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"sync/atomic"
	"testing"
	"wildberries-go-course/L0/model"
)

func TestTreePseudoLRUCachePutAndGet(t *testing.T) {
	// Given
	cache := TreePseudoLRUCache[string]{}

	// When
	cache.Put(1, "value1")
	cache.Put(2, "value2")
	cache.Put(3, "value3")

	// Then
	if val, found := cache.Get(1); !found || val != "value1" {
		t.Errorf("Expected 'value1', got '%s', found: %t", val, found)
	}

	if val, found := cache.Get(2); !found || val != "value2" {
		t.Errorf("Expected 'value2', got '%s', found: %t", val, found)
	}

	if val, found := cache.Get(3); !found || val != "value3" {
		t.Errorf("Expected 'value3', got '%s', found: %t", val, found)
	}

	if _, found := cache.Get(4); found {
		t.Errorf("Expected to NOT found the key 4, but found it")
	}
}

func TestTreePseudoLRUCacheCapacityLimit(t *testing.T) {
	// Given
	cache := TreePseudoLRUCache[uint64]{}

	// When
	for i := uint64(1); i < 67; i++ {
		cache.Put(i, i)
	}

	// Then
	if _, found := cache.Get(1); found {
		t.Error("Expected key 1 to be evicted")
	}

	if val, found := cache.Get(64); !found || val != 64 {
		t.Errorf("Expected key 64 to exist with value 64, got %d, found: %t", val, found)
	}
}

var seed = int64(131071)

func BenchmarkTreePseudoLRUCache_Put(b *testing.B) {
	cache := TreePseudoLRUCache[int]{}

	const numKeys = 120
	const amountOfRandomOperationKeys = 1000

	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(i)
	}

	randomIndices := make([]int, amountOfRandomOperationKeys)

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < amountOfRandomOperationKeys; i++ {
		randomIndices[i] = rng.Intn(numKeys)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		index := randomIndices[i%amountOfRandomOperationKeys]
		cache.Put(keys[index], i)
	}
}

func BenchmarkTreePseudoLRUCache_Get(b *testing.B) {
	const capacity = 64
	cache := TreePseudoLRUCache[int]{}

	const numKeys = 1200
	const amountOfRandomOperationKeys = 100000

	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(i)
		cache.Put(uint64(i), i)
	}

	randomIndices := make([]int, amountOfRandomOperationKeys)

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < amountOfRandomOperationKeys; i++ {
		randomIndices[i] = rng.Intn(numKeys)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		index := randomIndices[i%amountOfRandomOperationKeys]
		cache.Get(keys[index])

	}
}

func TestContainsSIMD(t *testing.T) {
	var ids [64]uint64
	ids[12] = 12
	ids[15] = 51
	ids[2] = 20
	var index int64
	index = ContainsSIMD(&ids, 12)
	if index != 12 {
		t.Errorf("index should be 12 but its %d", index)
	}

	index = ContainsSIMD(&ids, 51)
	if index != 15 {
		t.Errorf("index should be 15 but its %d", index)
	}

	index = ContainsSIMD(&ids, 20)
	if index != 2 {
		t.Errorf("index should be 2 but its %d", index)
	}

	index = ContainsSIMD(&ids, 3)
	if index != -1 {
		t.Errorf("index should be -1 but its %d", index)
	}

}

func BenchmarkTreePseudoLRUCache_Parallel(b *testing.B) {
	cache := ShardedTreeCache[*model.Order]{}

	const poolSize = 2048
	orders := make([]model.Order, poolSize)
	for i := 0; i < poolSize; i++ {
		orders[i] = model.Order{
			OrderUID:   fmt.Sprintf("UID-%06d", i),
			CustomerID: fmt.Sprintf("CUST-%d", i),
			Items:      []model.Item{{Name: "Test Item", Price: 1000}},
		}

		hashId := maphash.String(HashSeed, orders[i].OrderUID)
		cache.Put(hashId, &orders[i])

	}

	var threadCounter int64

	b.ResetTimer()

	// this for debug
	// var found atomic.Uint64
	// var lost atomic.Uint64
	b.RunParallel(func(pb *testing.PB) {
		localSeed := seed + atomic.AddInt64(&threadCounter, 1)
		rng := rand.New(rand.NewSource(localSeed))

		for pb.Next() {
			id := uint64(rng.Intn(poolSize))

			operation := rng.Float32()
			// 80% to Get, 20% to put
			if operation < 0.8 {
				// to do properly we want to use hash orderUID because it donst have a number id
				// but ideally you want to change Order struct and add id into it
				hashId := maphash.String(HashSeed, orders[id].OrderUID)
				// _, isFound := cache.Get(hashId)
				_, _ = cache.Get(hashId)
				// if isFound {
				// found.Add(1)
				// } else {
				// lost.Add(1)
				// }
			} else {

				hashId := maphash.String(HashSeed, orders[id].OrderUID)
				cache.Put(hashId, &orders[id])
			}
		}
	})
	// fmt.Printf("found :%v, lost :%v\n", found.Load(), lost.Load())
}
