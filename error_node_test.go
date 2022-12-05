package tower

import (
	"errors"
	"fmt"
	"github.com/kinbiko/jsonassert"
	"strings"
	"testing"
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
			if !strings.Contains(string(got), "tower/error_node_test.go") {
				t.Error("expected to see caller in tower/error_node_test.go")
			}
			if strings.Count(string(got), "tower/error_node_test.go") != 4 {
				t.Error("expected to see four callers field in tower/error_node_test.go")
			}
		})
	}
}
