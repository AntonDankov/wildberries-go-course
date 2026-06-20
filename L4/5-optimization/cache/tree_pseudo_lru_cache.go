package cache

import (
	"math/bits"
	simd "simd/archsimd"
	"sync"
	"unsafe"
	"wildberries-go-course/L0/model"
)

type TreePseudoLRUCache[T any] struct {
	ids     [64]uint64
	lruTree uint64

	values [64]T
	mutex  sync.Mutex
}

func (treeCache *TreePseudoLRUCache[T]) Get(id uint64) (T, bool) {
	treeCache.mutex.Lock()
	defer treeCache.mutex.Unlock()
	index := ContainsSIMD(&treeCache.ids, id)
	if index != -1 {
		treeCache.UpdateLRU(uint64(index))
		return treeCache.values[index], true
	}
	var empty T
	return empty, false
}

func (treeCache *TreePseudoLRUCache[T]) Put(id uint64, value T) {
	treeCache.mutex.Lock()
	defer treeCache.mutex.Unlock()

	containsIndex := ContainsSIMD(&treeCache.ids, id)
	if containsIndex == -1 {
		// we didnt found it
		indexLRU := treeCache.FindLeastRecent()
		treeCache.values[indexLRU] = value
		treeCache.ids[indexLRU] = id
		treeCache.UpdateLRU(indexLRU)
	} else {
		// update it as recent one
		treeCache.UpdateLRU(uint64(containsIndex))
	}
}

// we go backwards in our binary tree
// 0 is go left, 1 is go rigth
func (treeCache *TreePseudoLRUCache[T]) UpdateLRU(cacheIndex uint64) {
	currentLRU := treeCache.lruTree
	index := cacheIndex + 63

	for index > 0 {
		parent := (index - 1) / 2

		// isLeft will be 1 if it's odd, or 0 if it's even
		// and we point to the opposite so if it was left we point to right
		isLeft := index & 1

		// clear the bit and set it as isLeft
		currentLRU &^= (1 << parent)
		currentLRU |= (isLeft << parent)

		index = parent
	}
	treeCache.lruTree = currentLRU
}

func (treeCache *TreePseudoLRUCache[T]) FindLeastRecent() uint64 {
	index := uint64(0)

	currentLRU := treeCache.lruTree
	for range 6 {
		bit := (currentLRU & (1 << index)) != 0
		if !bit {
			index = index*2 + 1
		} else {
			index = index*2 + 2
		}

	}

	return index - 63

}

func ContainsSIMD(ids *[64]uint64, target uint64) int64 {
	vTarget := simd.BroadcastUint64x4(target)

	for i := 0; i < 16; i++ {
		chunk := (*[4]uint64)(unsafe.Pointer(&ids[i*4]))
		vChunk := simd.LoadUint64x4(chunk)

		mask := vChunk.Equal(vTarget).ToBits()
		if mask != 0 {
			return int64(i*4) + int64(bits.TrailingZeros8(mask))
		}
	}
	return -1
}

var GlobalShardedTreeCache = ShardedTreeCache[*model.Order]{}

type ShardedTreeCache[T any] struct {
	shards [16]TreePseudoLRUCache[T]
}

func (s *ShardedTreeCache[T]) Get(id uint64) (T, bool) {
	shardIndex := id & 15
	return s.shards[shardIndex].Get(id)
}

func (s *ShardedTreeCache[T]) Put(id uint64, value T) {
	shardIndex := id & 15
	s.shards[shardIndex].Put(id, value)
}
