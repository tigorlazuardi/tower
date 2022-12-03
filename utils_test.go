package tower

import (
	"reflect"
	"testing"
)

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
