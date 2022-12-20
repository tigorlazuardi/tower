package tower

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

type mockLogger struct {
	called bool
}

func newMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Log(ctx context.Context, entry Entry) {
	m.called = true
}

func (m *mockLogger) LogError(ctx context.Context, err Error) {
	m.called = true
}

func newMockMessenger(count int) *mockMessenger {
	wg := &sync.WaitGroup{}
	wg.Add(count)
	return &mockMessenger{
		wg: wg,
	}
}

type mockMessenger struct {
	called bool
	wg     *sync.WaitGroup
}

func (m *mockMessenger) Name() string {
	return "mock"
}

func (m *mockMessenger) SendMessage(ctx context.Context, msg MessageContext) {
	m.called = true
	m.wg.Done()
}

func (m *mockMessenger) Wait(ctx context.Context) error {
	m.wg.Wait()
	return nil
}

func TestDbg(t *testing.T) {
	type args struct {
		a any
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "output is expected",
			args: args{
				a: "foo",
			},
		},
		{
			name: "output is expected",
			args: args{
				a: map[string]any{
					"foo": "bar",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			Dbg(tt.args.a)
		})
	}
}

func TestCast(t *testing.T) {
	type args[T any] struct {
		in []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []any
	}
	tests := []testCase[int]{
		{
			name: "output is expected",
			args: args[int]{
				in: []int{1, 2, 3},
			},
			want: []any{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Cast(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cast() = %v, want %v", got, tt.want)
			}
		})
	}
}
