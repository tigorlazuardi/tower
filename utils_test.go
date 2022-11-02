package tower

import "testing"

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
