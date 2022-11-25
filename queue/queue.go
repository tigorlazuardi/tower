package queue

import (
	"sync"
	"sync/atomic"
)

// Queue is a lock-free unbounded queue.
type Queue[T any] struct {
	head  *node[T]
	tail  *node[T]
	hlock *sync.Mutex
	tlock *sync.Mutex
	len   uint64
}

type node[T any] struct {
	value *T
	next  *node[T]
}

// New returns an empty concurrent safe queue.
func New[T any]() *Queue[T] {
	n := &node[T]{}
	return &Queue[T]{head: n, tail: n, hlock: &sync.Mutex{}, tlock: &sync.Mutex{}}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *Queue[T]) Enqueue(v T) {
	n := &node[T]{value: &v}
	q.tlock.Lock()
	q.tail.next = n
	q.tail = n
	q.tlock.Unlock()
	atomic.AddUint64(&q.len, 1)
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if empty.
func (q *Queue[T]) Dequeue() T {
	q.hlock.Lock()
	n := q.head
	newHead := n.next
	if newHead == nil {
		q.hlock.Unlock()
		var t T
		return t
	}
	v := *newHead.value
	newHead.value = nil
	q.head = newHead
	q.hlock.Unlock()
	atomic.AddUint64(&q.len, ^uint64(0))
	return v
}

// Len Returns the current length of queue.
func (q *Queue[T]) Len() uint64 {
	return atomic.LoadUint64(&q.len)
}
