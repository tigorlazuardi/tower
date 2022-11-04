package queue

import (
	"sync/atomic"
	"unsafe"
)

// Queue is a lock-free unbounded queue.
type Queue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	len  uint64
}
type node[T any] struct {
	value T
	next  unsafe.Pointer
}

// New returns an empty multi threaded safe queue.
func New[T any]() *Queue[T] {
	n := unsafe.Pointer(&node[T]{})
	return &Queue[T]{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *Queue[T]) Enqueue(v T) {
	n := &node[T]{value: v}
	for {
		tail := load[T](&q.tail)
		next := load[T](&tail.next)
		if tail == load[T](&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					cas(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					atomic.AddUint64(&q.len, 1)
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns zero value of T if nil.
func (q *Queue[T]) Dequeue() T {
	for {
		head := load[T](&q.head)
		tail := load[T](&q.tail)
		next := load[T](&head.next)
		if head == load[T](&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					var t T
					return t
				}
				// tail is falling behind.  try to advance it
				cas(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas(&q.head, head, next) {
					atomic.AddUint64(&q.len, ^uint64(0))
					return v // Dequeue is done.  return
				}
			}
		}
	}
}

// Returns the current length of queue.
func (q *Queue[T]) Len() uint64 {
	return atomic.LoadUint64(&q.len)
}

func load[T any](p *unsafe.Pointer) (n *node[T]) {
	return (*node[T])(atomic.LoadPointer(p))
}

func cas[T any](p *unsafe.Pointer, old, new *node[T]) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
