package tower

import "testing"

func TestCaller_ShortOrigin(t *testing.T) {
	tests := []struct {
		name string
		c    Caller
		want string
	}{
		{
			name: "test",
			c:    func() Caller { c, _ := GetCaller(1); return c }(),
			want: "github.com/tigorlazuardi/tower.TestCaller_ShortOrigin.func1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.ShortOrigin(); got != tt.want {
				t.Errorf("Caller.ShortOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
