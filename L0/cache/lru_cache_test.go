package cache

import (
	"testing"
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
