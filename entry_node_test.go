package tower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kinbiko/jsonassert"
	"reflect"
	"testing"
	"time"
)

func TestEntryNode(t *testing.T) {
	tow := NewTower(Service{Name: "test"})
	now := time.Now()
	caller := GetCaller(1)
	builder := tow.NewEntry("foo")
	builder.Code(600).
		Level(ErrorLevel).
		Key("foo").
		Time(now).
		Caller(caller)

	node := builder.Freeze()
	if node == nil {
		t.Fatal("Expected entry node to be non-nil")
	}
	if node.Code() != 600 {
		t.Errorf("Expected entry node code to be 200, got %d", node.Code())
	}
	if node.HTTPCode() != 200 {
		t.Errorf("Expected entry node http code to be 200, got %d", node.HTTPCode())
	}
	if node.Service() != (Service{Name: "test"}) {
		t.Errorf("Expected entry node service to be test, got %s", node.Service())
	}
	builder.Code(500)
	node = builder.Freeze()
	if node.HTTPCode() != 500 {
		t.Errorf("Expected entry node http code to be 500, got %d", node.HTTPCode())
	}
	builder.Code(1301)
	node = builder.Freeze()
	if node.HTTPCode() != 301 {
		t.Errorf("Expected entry node http code to be 301, got %d", node.HTTPCode())
	}
	if node.Level() != ErrorLevel {
		t.Errorf("Expected entry node level to be ErrorLevel, got %s", node.Level())
	}
	if node.Key() != "foo" {
		t.Errorf("Expected entry node key to be foo, got %s", node.Key())
	}
	if node.Message() != "foo" {
		t.Errorf("Expected entry node message to be foo, got %s", node.Message())
	}
	if !node.Time().Equal(now) {
		t.Errorf("Expected entry node time to be %s, got %s", now, node.Time())
	}
	if !reflect.DeepEqual(node.Caller(), caller) {
		t.Errorf("Expected entry node caller to be %v, got %v", caller, node.Caller())
	}

	builder.Context(1)
	node = builder.Freeze()
	if len(node.Context()) != 1 {
		t.Fatalf("Expected entry node context to be 1, got %d", len(node.Context()))
	}
	j := jsonassert.New(t)
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Expected entry node to marshal to JSON without error, got %v", err)
	}
	j.Assertf(string(b), `
		{
			"time": "<<PRESENCE>>",
			"code": 1301,
			"message": "foo",
			"caller": "<<PRESENCE>>",
			"key": "foo",
			"level": "error",
			"service": {"name": "test"},
			"context": 1
		}`,
	)
	if t.Failed() {
		out := new(bytes.Buffer)
		_ = json.Indent(out, b, "", "    ")
		fmt.Println(out.String())
	}
}
