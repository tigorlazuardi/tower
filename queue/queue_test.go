package queue_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tigorlazuardi/tower-go/queue"
)

func TestLockFreeQueue(t *testing.T) {
	queue := queue.New[int]()
	count := uint64(0)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(5000)
	for i := 1; i <= 5000; i++ {
		queue.Enqueue(i)
		go func() {
			<-ctx.Done()
			j := queue.Dequeue()
			if j == 0 {
				t.Error("unexpected 0 value from queue. there should be no 0 value.")
			}
			atomic.AddUint64(&count, 1)
			wg.Done()
		}()
	}
	if queue.Len() != 5000 {
		t.Errorf("expected queue to have 5000 length, but got %d length", queue.Len())
	}
	cancel()
	wg.Wait()
	if queue.Len() != 0 {
		t.Errorf("expected queue to have 0 length, but got %d length", queue.Len())
	}
	if count != 5000 {
		t.Errorf("expected count to be 5000, but got %d", count)
	}
}

func BenchmarkLockFreeQueue(b *testing.B) {
	queue := queue.New[int]()
	wg := sync.WaitGroup{}
	wg.Add(b.N * 2)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			queue.Enqueue(i)
			wg.Done()
		}(i)
		go func() {
			queue.Dequeue()
			wg.Done()
		}()
	}
	wg.Wait()
}
