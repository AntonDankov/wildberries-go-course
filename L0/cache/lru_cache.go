package cache

import (
	"bytes"
	"fmt"
	"sync"
	"wildberries-go-course/L0/model"
)

var GlobalOrderCache = NewSyncLRUCache[model.Order](100)

type SyncLRUCache[T any] struct {
	linkedList *LinkedList[T]
	hashMap    map[string]*Node[T]
	capacity   int
	mutex      sync.Mutex
}

func NewSyncLRUCache[T any](capacity int) *SyncLRUCache[T] {
	return &SyncLRUCache[T]{
		linkedList: NewLinkedList[T](),
		hashMap:    make(map[string]*Node[T]),
		capacity:   capacity,
		mutex:      sync.Mutex{},
	}
}

func (cache *SyncLRUCache[T]) Get(key string) (T, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	node, exist := cache.hashMap[key]
	var empty T
	if !exist {
		return empty, false
	}

	cache.linkedList.MoveToTail(node)

	return node.Value, true
}

func (cache *SyncLRUCache[T]) Put(key string, value T) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	node, exist := cache.hashMap[key]
	if exist {
		node.Value = value

		return
	} else {

		list := cache.linkedList

		if list.Size == cache.capacity {
			cache.RemoveOldest()
		}

		node = list.Add(key, value)
		cache.hashMap[key] = node
	}
}

func (cache *SyncLRUCache[T]) PrintAllNodes() string {
	var buffer bytes.Buffer
	cur := cache.linkedList.Head.Next
	for cur != nil {
		buffer.WriteString(fmt.Sprint(cur.Value))
		cur = cur.Next
		if cur != nil {
			buffer.WriteString("->")
		}
	}

	return buffer.String()
}

type LinkedList[T any] struct {
	Head *Node[T]
	Tail *Node[T]
	Size int
}

func NewLinkedList[T any]() *LinkedList[T] {
	dummy := &Node[T]{
		Next: nil,
		Prev: nil,
	}
	return &LinkedList[T]{
		Head: dummy,
		Tail: dummy,
		Size: 0,
	}
}

func (list *LinkedList[T]) Add(key string, value T) *Node[T] {
	node := &Node[T]{
		Key:   key,
		Value: value,
		Next:  nil,
		Prev:  list.Tail,
	}
	list.Tail.Next = node
	list.Tail = node
	list.Size++

	return node
}

func (cache *SyncLRUCache[T]) RemoveOldest() {
	list := cache.linkedList
	if list.Size == 0 {
		return
	}
	oldest := list.Head.Next
	list.Head.Next = oldest.Next
	if oldest.Next != nil {
		oldest.Next.Prev = list.Head
	}
	list.Size--
	delete(cache.hashMap, oldest.Key)
}

func (list *LinkedList[T]) MoveToTail(node *Node[T]) {
	last := list.Tail

	if node != last {
		prev := node.Prev
		prev.Next = node.Next
		node.Next.Prev = prev

		last.Next = node
		node.Prev = last
		node.Next = nil
		list.Tail = node
	}
}

type Node[T any] struct {
	Key   string
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}
