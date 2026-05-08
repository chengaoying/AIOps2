package queue

import (
	"container/ring"
	"sync"
)

type MemoryQueue struct {
	maxSize int
	data    *ring.Ring
	size    int
	mu      sync.Mutex
	head    *ring.Ring
}

func New(maxSize int) *MemoryQueue {
	r := ring.New(maxSize)
	return &MemoryQueue{
		maxSize: maxSize,
		data:    r,
		head:    r,
	}
}

func (q *MemoryQueue) Enqueue(item any) any {
	q.mu.Lock()
	defer q.mu.Unlock()

	evicted := q.head.Value
	q.head.Value = item
	q.head = q.head.Next()
	q.size++

	if q.size > q.maxSize {
		q.size = q.maxSize
		oldest := evicted
		evicted = oldest.Value
		oldest.Value = nil
		return evicted
	}
	return nil
}

func (q *MemoryQueue) Dequeue() (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size == 0 {
		return nil, false
	}

	item := q.head.Value
	q.head.Value = nil
	q.head = q.head.Next()
	q.size--
	return item, true
}

func (q *MemoryQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

func (q *MemoryQueue) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size >= q.maxSize
}
