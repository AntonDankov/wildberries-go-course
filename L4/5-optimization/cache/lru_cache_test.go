package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"testing"
	"wildberries-go-course/L0/model"
)

func TestNewSyncLRUCache(t *testing.T) {
	cache := NewSyncLRUCache[int](5)
	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}

	if cache.capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", cache.capacity)
	}

	if cache.linkedList.Size != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.linkedList.Size)
	}

	if len(cache.hashMap) != 0 {
		t.Errorf("Expected empty hashMap, got size %d", len(cache.hashMap))
	}
}

func TestPutAndGet(t *testing.T) {
	// Given
	cache := NewSyncLRUCache[string](3)

	// When
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	// Then
	if val, found := cache.Get("key1"); !found || val != "value1" {
		t.Errorf("Expected 'value1', got '%s', found: %t", val, found)
	}

	if val, found := cache.Get("key2"); !found || val != "value2" {
		t.Errorf("Expected 'value2', got '%s', found: %t", val, found)
	}

	if val, found := cache.Get("key3"); !found || val != "value3" {
		t.Errorf("Expected 'value3', got '%s', found: %t", val, found)
	}
}

func TestUpdateExistingKey(t *testing.T) {
	// Given
	cache := NewSyncLRUCache[int](3)
	// When
	cache.Put("key1", 100)
	cache.Put("key1", 200)

	// Then
	if val, found := cache.Get("key1"); !found || val != 200 {
		t.Errorf("Expected updated value 200, got %d, found: %t", val, found)
	}

	if cache.linkedList.Size != 1 {
		t.Errorf("Expected size 1 after update, got %d", cache.linkedList.Size)
	}
}

func TestCapacityLimit(t *testing.T) {
	// Given
	cache := NewSyncLRUCache[int](2)

	// When
	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	// Then
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}

	if val, found := cache.Get("key2"); !found || val != 2 {
		t.Errorf("Expected key2 to exist with value 2, got %d, found: %t", val, found)
	}
	if val, found := cache.Get("key3"); !found || val != 3 {
		t.Errorf("Expected key3 to exist with value 3, got %d, found: %t", val, found)
	}

	if cache.linkedList.Size != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.linkedList.Size)
	}
}

func TestEmptyCache(t *testing.T) {
	// Given
	cache := NewSyncLRUCache[string](5)

	// When
	cache.RemoveOldest()

	// Then
	if _, found := cache.Get("any"); found {
		t.Error("Expected no keys in empty cache")
	}

	if cache.linkedList.Size != 0 {
		t.Errorf("Expected size 0 after RemoveOldest on empty cache, got %d", cache.linkedList.Size)
	}
}

func TestComplexScenario(t *testing.T) {
	// Given
	cache := NewSyncLRUCache[int](3)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)
	cache.Get("key2")
	cache.Put("key4", 4)

	if cache.linkedList.Size != 3 {
		t.Errorf("Expected cache size 2, got %d", cache.linkedList.Size)
	}
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}

	cache.Put("key1", 1)
	// Then
	if _, found := cache.Get("key3"); found {
		t.Error("Expected key3 to be evicted")
	}
	if _, found := cache.Get("key2"); !found {
		t.Error("Expected key2 to be in cache")
	}

	if _, found := cache.Get("key1"); !found {
		t.Error("Expected key1 to be in cache")
	}
	if _, found := cache.Get("key4"); !found {
		t.Error("Expected key4 to be in cache")
	}

	if cache.linkedList.Size != 3 {
		t.Errorf("Expected cache size 2, got %d", cache.linkedList.Size)
	}
}

func BenchmarkCache_Put(b *testing.B) {
	const capacity = 64
	cache := NewSyncLRUCache[int](capacity)

	seed := int64(131071)
	const numKeys = 120
	const amountOfRandomOperationKeys = 1000

	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = strconv.Itoa(i)
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

func BenchmarkCache_Get(b *testing.B) {
	const capacity = 64
	cache := NewSyncLRUCache[int](capacity)

	const numKeys = 120
	const amountOfRandomOperationKeys = 100000

	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = strconv.Itoa(i)
		cache.Put(keys[i], i)
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

func BenchmarkCache_Parallel(b *testing.B) {
	cache := NewSharedCache[*model.Order](64)

	// 1. Pre-allocate data to avoid benchmarking the Garbage Collector
	const poolSize = 24092 // We have 64 cache slots, 256 items guarantees heavy eviction contention
	orders := make([]model.Order, poolSize)
	for i := 0; i < poolSize; i++ {
		orders[i] = model.Order{
			OrderUID:   fmt.Sprintf("UID-%06d", i),
			CustomerID: fmt.Sprintf("CUST-%d", i),
			Items:      []model.Item{{Name: "Test Item", Price: 1000}},
		}
		cache.Put(orders[i].OrderUID, &orders[i])
	}

	// We use an atomic counter to ensure each thread gets a uniquely seeded PRNG
	var threadCounter int64

	b.ResetTimer() // Reset timer after setup

	// var found atomic.Uint64
	// var lost atomic.Uint64
	// RunParallel launches multiple Goroutines and runs the loop simultaneously
	b.RunParallel(func(pb *testing.PB) {
		// Create a local, lock-free random generator for THIS thread
		// We add the threadCounter to the base seed so threads don't perfectly overlap
		localSeed := seed + atomic.AddInt64(&threadCounter, 1)
		rng := rand.New(rand.NewSource(localSeed))

		for pb.Next() {
			// Generate a random ID between 1 and 256
			// (We add 1 because 0 is often a default zero-value, and we don't want to conflict with Tombstones)
			id := (rng.Intn(poolSize))
			uuid := orders[id].OrderUID
			// idsrt := strconv.Itoa(id)

			// Simulate real-world workload: 80% Reads (Get), 20% Writes (Put)
			operation := rng.Float32()

			if operation < 0.8 {
				// 80% GET
				_, _ = cache.Get(uuid)
				// if isFound {
				// 	found.Add(1)
				// } else {
				// 	lost.Add(1)
				// }
			} else {
				// 20% PUT
				// We pass the pre-allocated Order by value (it copies the struct fields,
				// but because slices are references, the Items array doesn't copy deeply).
				cache.Put(uuid, &orders[id])
			}
		}
	})

	// fmt.Printf("found :%v, lost :%v\n", found.Load(), lost.Load())
}
