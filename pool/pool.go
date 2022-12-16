package pool

import (
	"errors"
	"sync"
)

// Pool is a wrapper around sync.Pool that allows for generic.
type Pool[T any] struct {
	inner *sync.Pool
}

// Get selects an arbitrary item of T from the Pool, removes it from the Pool, and returns it to the caller. Get may choose to
// ignore the pool and treat it as empty. Callers should not assume any relation between values passed to Put and the values
// returned by Get.
func New[T any](generator func() T) *Pool[T] {
	if generator == nil {
		panic(errors.New("generator function cannot be nil"))
	}
	return &Pool[T]{
		inner: &sync.Pool{
			New: func() any {
				return generator()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.inner.Get().(T) //nolint
}

func (p *Pool[T]) Put(t T) {
	p.inner.Put(t)
}
