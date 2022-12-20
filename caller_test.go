package tower

import (
	"bytes"
	"strings"
	"testing"
)

func TestCaller(t *testing.T) {
	c := GetCaller(1)
	if c == nil {
		t.Fatal("Expected caller to be non-nil")
	}
	if c.Name() != "github.com/tigorlazuardi/tower.TestCaller" {
		t.Errorf("Expected caller name to be github.com/tigorlazuardi/tower.TestCaller, got %s", c.Name())
	}
	if c.ShortName() != "tower.TestCaller" {
		t.Errorf("Expected caller short name to be tower.TestCaller, got %s", c.ShortName())
	}
	if !strings.Contains(c.File(), "tower/caller_test.go") {
		t.Errorf("Expected caller file to be tower/caller_test.go, got %s", c.File())
	}
	if c.PC() == 0 {
		t.Errorf("Expected caller pc to be non-zero")
	}
	if !strings.Contains(c.FormatAsKey(), "tower_caller_test.go_") {
		t.Errorf("Expected caller format as key to be tower_caller_test.go_, got %s", c.FormatAsKey())
	}
	if c.Line() <= 2 {
		t.Errorf("expected caller line to be greater than 2, got %d", c.Line())
	}
	if c.Function() == nil {
		t.Errorf("Expected caller function to be non-nil")
	}
	if !strings.Contains(c.String(), "tower/caller_test.go:") {
		t.Errorf("Expected caller string to be tower/caller_test.go:, got %s", c.String())
	}
	if !strings.Contains(c.ShortSource(), "tower/caller_test.go") {
		t.Errorf("Expected caller short source to be tower/caller_test.go, got %s", c.ShortSource())
	}
	cal := c.(*caller)
	b, err := cal.MarshalJSON()
	if err != nil {
		t.Errorf("Expected caller marshal json to be nil, got %s", err)
	}
	if bytes.Contains(b, []byte("tower/caller_test.go:9")) {
		t.Errorf("Expected caller marshal json to be tower/caller_test.go:, got %s", b)
	}
}
