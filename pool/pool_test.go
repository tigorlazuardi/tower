package pool

import (
	"testing"
)

func TestPool(t *testing.T) {
	pool := New(func() *int {
		w := 2
		return &w
	})

	got := pool.Get()

	if got == nil {
		t.Errorf("got nil, want non-nil")
		return
	}

	if *got != 2 {
		t.Errorf("got %d, want 2", *got)
		return
	}

	*got = 3
	pool.Put(got)

	got = pool.Get()
	if *got != 3 {
		t.Errorf("got %d, want 3", *got)
		return
	}
	got2 := pool.Get()
	if got2 == nil {
		t.Errorf("got nil, want not-nil")
		return
	}
	if *got2 != 2 {
		t.Errorf("got %d, want 2", *got2)
		return
	}
}

func TestPoolNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should have panicked")
		}
	}()

	New[any](nil)
}
