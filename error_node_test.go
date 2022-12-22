package tower

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kinbiko/jsonassert"
	"strings"
	"testing"
	"time"
)

func TestErrorNode_CodeBlockJSON(t *testing.T) {
	tests := []struct {
		name      string
		baseError error
		messages  []string
		want      string
		wantErr   bool
	}{
		{
			name:      "expected output",
			baseError: errors.New("base error"),
			messages:  []string{"message 1", "message 2", "message 3"},
			want: `
{
   "message": "message 3",
   "caller": "<<PRESENCE>>",
   "error": {
      "message": "message 2",
      "caller": "<<PRESENCE>>",
      "error": {
         "message": "message 1",
         "caller": "<<PRESENCE>>",
         "error": {
            "time": "<<PRESENCE>>",
            "code": 500,
            "message": "base error",
            "caller": "<<PRESENCE>>",
            "level": "error",
            "service": {
               "name": "testing-code-block",
               "environment": "testing",
               "type": "unit-test"
            },
            "error": {
               "summary": "base error"
            }
         }
      }
   }
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tow := NewTower(Service{
				Name:        "testing-code-block",
				Environment: "testing",
				Type:        "unit-test",
			})
			err := tow.Wrap(tt.baseError).Freeze()
			for _, e := range tt.messages {
				err = tow.WrapFreeze(err, e)
			}
			got, errCB := err.(*ErrorNode).CodeBlockJSON()
			if (errCB != nil) != tt.wantErr {
				t.Errorf("ErrorNode.CodeBlockJSON() error = %v, wantErr %v", errCB, tt.wantErr)
				return
			}
			j := jsonassert.New(t)
			j.Assertf(string(got), tt.want)
			if t.Failed() {
				fmt.Println(string(got))
			}
			if !strings.Contains(string(got), "error_node_test.go") {
				t.Error("expected to see caller in error_node_test.go")
			}
			if strings.Count(string(got), "error_node_test.go") != 4 {
				t.Error("expected to see four callers field in error_node_test.go")
			}
		})
	}
}

type mockImplError struct{}

func (m mockImplError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"message": m.Message(),
		"caller":  m.Caller(),
		"code":    m.Code(),
		"error":   m.Error(),
		"level":   m.Level(),
		"service": m.Service(),
		"time":    m.Time(),
	})
}

func (m mockImplError) Error() string {
	return "mock"
}

func (m mockImplError) Caller() Caller {
	return GetCaller(1)
}

func (m mockImplError) Code() int {
	return 500
}

func (m mockImplError) Context() []any {
	return nil
}

func (m mockImplError) Unwrap() error {
	return nil
}

func (m mockImplError) WriteError(w LineWriter) {}

func (m mockImplError) HTTPCode() int {
	return 500
}

func (m mockImplError) Key() string {
	return ""
}

func (m mockImplError) Level() Level {
	return ErrorLevel
}

func (m mockImplError) Message() string {
	return "mocking time"
}

func (m mockImplError) Time() time.Time {
	return time.Now()
}

func (m mockImplError) Service() Service {
	return Service{
		Name: "mock",
	}
}

func (m mockImplError) Log(ctx context.Context) Error {
	return m
}

func (m mockImplError) Notify(ctx context.Context, opts ...MessageOption) Error {
	return m
}

func TestErrorNode_MarshalJSON(t *testing.T) {
	tow := NewTower(Service{Name: "test"})
	tests := []struct {
		name    string
		err     *ErrorNode
		want    string
		wantErr bool
	}{
		{
			name: "expected output - simple error",
			err:  tow.Wrap(errors.New("base error")).Message("bar").Freeze().(*ErrorNode),
			want: `
				{
				   "time": "<<PRESENCE>>",
				   "code": 500,
				   "message": "bar",
				   "caller": "<<PRESENCE>>",
				   "level": "error",
				   "service": {
					  "name": "test"
				   },
				   "error": {
					  "summary": "base error"
				   }
				}`,
			wantErr: false,
		},
		{
			name: "expected output - nested error",
			err: func() *ErrorNode {
				base := tow.Wrap(errors.New("base error")).Message("bar").Freeze()
				return tow.WrapFreeze(base, "error 1").(*ErrorNode)
			}(),
			want: `
				{
				   "time": "<<PRESENCE>>",
				   "code": 500,
				   "message": "error 1",
				   "caller": "<<PRESENCE>>",
				   "level": "error",
				   "service": {
					  "name": "test"
				   },
				   "error": {
					  "message": "bar",
					  "caller": "<<PRESENCE>>",
					  "error": {
						 "summary": "base error"
					  }
				   }
				}`,
			wantErr: false,
		},
		{
			name: "expected output - nested error with wrap that does nothing",
			err: func() *ErrorNode {
				base := tow.Wrap(errors.New("base error")).Message("bar").Freeze()
				base = tow.Wrap(base).Freeze()
				base = tow.Wrap(base).Freeze()
				return tow.WrapFreeze(base, "error 1").(*ErrorNode)
			}(),
			want: `
				{
				   "time": "<<PRESENCE>>",
				   "code": 500,
				   "message": "error 1",
				   "caller": "<<PRESENCE>>",
				   "level": "error",
				   "service": {
					  "name": "test"
				   },
				   "error": {
					  "message": "bar",
					  "caller": "<<PRESENCE>>",
					  "error": {
						 "summary": "base error"
					  }
				   }
				}`,
			wantErr: false,
		},
		{
			name: "expected output - wrap other Error implementation",
			err: func() *ErrorNode {
				base := tow.Wrap(mockImplError{}).Code(400).Freeze()
				return tow.WrapFreeze(base, "error 1").(*ErrorNode)
			}(),
			want: `
				{
				   "time": "<<PRESENCE>>",
				   "code": 400,
				   "message": "error 1",
				   "caller": "<<PRESENCE>>",
				   "level": "error",
				   "service": {
					  "name": "test"
				   },
				   "error": {
					  "caller": "<<PRESENCE>>",
					  "code": 500,
					  "error": "mock",
					  "level": 2,
					  "message": "mocking time",
					  "service": {
						 "name": "mock"
					  },
					  "time": "<<PRESENCE>>"
				   }
				}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.err
			got, err := e.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			j := jsonassert.New(t)
			j.Assertf(string(got), tt.want)
			if t.Failed() {
				out := new(bytes.Buffer)
				_ = json.Indent(out, got, "", "   ")
				fmt.Println(out.String())
			}
		})
	}
}
