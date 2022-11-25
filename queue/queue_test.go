package queue_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tigorlazuardi/tower/queue"
)

func TestLockFreeQueue(t *testing.T) {
	q := queue.New[int]()
	count := uint64(0)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(5000)
	for i := 1; i <= 5000; i++ {
		q.Enqueue(i)
		go func() {
			<-ctx.Done()
			j := q.Dequeue()
			if j == 0 {
				t.Error("unexpected 0 value from queue. there should be no 0 value.")
			}
			atomic.AddUint64(&count, 1)
			wg.Done()
		}()
	}
	if q.Len() != 5000 {
		t.Errorf("expected queue to have 5000 length, but got %d length", q.Len())
	}
	cancel()
	wg.Wait()
	if q.Len() != 0 {
		t.Errorf("expected queue to have 0 length, but got %d length", q.Len())
	}
	if count != 5000 {
		t.Errorf("expected count to be 5000, but got %d", count)
	}
}

func BenchmarkLockFreeQueue(b *testing.B) {
	q := queue.New[int]()
	wg := sync.WaitGroup{}
	wg.Add(b.N * 2)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.Enqueue(i)
			wg.Done()
		}(i)
		go func() {
			q.Dequeue()
			wg.Done()
		}()
	}
	wg.Wait()
}
