package queue

import (
	"sync"
	"sync/atomic"
)

// Multi threaded safe implementation for FIFO Queue.
type Queue[T any] struct {
	head  *node
	tail  *node
	len   uint64
	hlock sync.Mutex
	tlock sync.Mutex
}

type node struct {
	value any
	next  *node
}

// Creates a new FIFO Queue.
func NewQueue[T any]() *Queue[T] {
	dummy := &node{}
	return &Queue[T]{head: dummy, tail: dummy}
}

// Adds a value to the tail of the queue.
func (q *Queue[T]) Enqueue(v T) {
	n := &node{value: v}
	q.tlock.Lock()
	q.tail.next = n
	q.tail = n
	q.tlock.Unlock()
	atomic.AddUint64(&q.len, 1)
}

// Dequeue removes and returns the value at the head of the queue.
// Returns zero value of T if the queue is empty.
func (q *Queue[T]) Dequeue() T {
	q.hlock.Lock()
	n := q.head
	newHead := n.next
	if newHead == nil {
		q.hlock.Unlock()
		var t T
		return t
	}
	v := newHead.value.(T) //nolint guaranteed to be safe from panics because it's already guarded by generics.
	newHead.value = nil
	q.head = newHead
	q.hlock.Unlock()
	atomic.AddUint64(&q.len, ^uint64(0))
	return v
}

func (q *Queue[T]) Len() uint64 {
	return atomic.LoadUint64(&q.len)
}
