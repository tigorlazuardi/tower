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

func Test_Error(t *testing.T) {
	tow := NewTower(Service{Name: "test"})
	l := newMockLogger()
	tow.SetLogger(l)
	m := newMockMessenger(1)
	tow.RegisterMessenger(m)
	base := errors.New("based")
	builder := tow.Wrap(base).Message("message 1")
	err := builder.Freeze()
	if err.Error() != "message 1: based" {
		t.Errorf("Error.Error() = %v, want %v", err.Error(), "message 1")
	}
	if err.Message() != "message 1" {
		t.Errorf("Error.Message() = %v, want %v", err.Message(), "message 1")
	}
	if err.Code() != 500 {
		t.Errorf("Error.Code() = %v, want %v", err.Code(), 500)
	}
	if err.HTTPCode() != 500 {
		t.Errorf("Error.HTTPCode() = %v, want %v", err.HTTPCode(), 500)
	}
	if errors.Unwrap(err) != base {
		t.Errorf("Error.Unwrap() = %v, want %v", errors.Unwrap(err), base)
	}
	ctx := context.Background()
	_ = err.Log(ctx).Notify(ctx)
	if !l.called {
		t.Error("Expected logger to be called")
	}
	err2 := m.Wait(ctx)
	if err2 != nil {
		t.Fatalf("Expected messenger to wait without error, got %v", err2)
	}
	if !m.called {
		t.Error("Expected messenger to be called")
	}

	b, errMarshal := json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	defer func() {
		if t.Failed() {
			out := new(bytes.Buffer)
			_ = json.Indent(out, b, "", "  ")
			t.Log(out.String())
		}
	}()
	j := jsonassert.New(t)
	j.Assertf(string(b), `
	{
		"time": "<<PRESENCE>>",
		"code": 500,
		"message": "message 1",
		"caller": "<<PRESENCE>>",
		"level": "error",
		"service": {"name": "test"},
		"error": {"summary": "based"}
	}`)

	now := time.Now()
	base2 := errors.New("based 2")
	builder.Level(WarnLevel).
		Caller(GetCaller(1)).
		Code(600).
		Error(base2).
		Context(1).
		Key("foo").
		Time(now)

	err = builder.Freeze()
	if err.HTTPCode() != 500 {
		t.Errorf("Error.HTTPCode() = %v, want %v", err.HTTPCode(), 500)
	}
	builder.Code(1400)
	err = builder.Freeze()
	if err.HTTPCode() != 400 {
		t.Errorf("Error.HTTPCode() = %v, want %v", err.HTTPCode(), 400)
	}
	b, errMarshal = json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	j.Assertf(string(b), `
	{
		"time": "%s",
		"code": 1400,
		"message": "message 1",
		"caller": "<<PRESENCE>>",
		"key": "foo",
		"level": "warn",
		"service": {"name": "test"},
		"context": 1,
		"error": {"summary": "based 2"}
	}`, now.Format(time.RFC3339))

	builder.Context(2).Message("foo %s", "bar").Key("foo %s", "bar")
	err = builder.Freeze()
	b, errMarshal = json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	j.Assertf(string(b), `
	{
		"time": "%s",
		"code": 1400,
		"message": "foo bar",
		"caller": "<<PRESENCE>>",
		"key": "foo bar",
		"level": "warn",
		"service": {"name": "test"},
		"context": [1, 2],
		"error": {"summary": "based 2"}
	}`, now.Format(time.RFC3339))

	l2 := newMockLogger()
	m2 := newMockMessenger(1)
	tow.messengers = make(map[string]Messenger)
	tow.RegisterMessenger(m2)
	tow.SetLogger(l2)
	_ = builder.Log(ctx)
	if !l2.called {
		t.Error("Expected logger to be called")
	}
	_ = builder.Notify(ctx)
	err3 := m2.Wait(ctx)
	if err3 != nil {
		t.Fatalf("Expected messenger to wait without error, got %v", err3)
	}
	if !m2.called {
		t.Error("Expected messenger to be called")
	}
}

type mockError struct {
	Message string
}

func (m mockError) Error() string {
	return m.Message
}

type mockJSONError struct{}

func (m mockJSONError) Error() string {
	return "mock"
}

func (m mockJSONError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"foo": "bar"})
}

type funcError func()

func (f funcError) Error() string {
	return "mock"
}

func TestError_WrapMarshalJSON(t *testing.T) {
	tow := NewTower(Service{Name: "test"})
	base := mockError{Message: "based"}
	builder := tow.Wrap(base).Message("message 1")
	err := builder.Freeze()
	b, errMarshal := json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	defer func() {
		if t.Failed() {
			out := new(bytes.Buffer)
			_ = json.Indent(out, b, "", "    ")
			t.Log(out.String())
		}
	}()
	j := jsonassert.New(t)
	j.Assertf(string(b), `
	{
		"time": "<<PRESENCE>>",
		"code": 500,
		"message": "message 1",
		"caller": "<<PRESENCE>>",
		"level": "error",
		"service": {"name": "test"},
		"error": {"summary": "based", "details": {"Message": "based"}}
	}`)
	builder.Error(mockJSONError{})
	err = builder.Freeze()
	b, errMarshal = json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	j.Assertf(string(b), `
	{
		"time": "<<PRESENCE>>",
		"code": 500,
		"message": "message 1",
		"caller": "<<PRESENCE>>",
		"level": "error",
		"service": {"name": "test"},
		"error": {"foo": "bar"}
	}`)
	builder.Error(funcError(func() {}))
	err = builder.Freeze()
	b, errMarshal = json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("Expected error to marshal to JSON without error, got %v", errMarshal)
	}
	j.Assertf(string(b), `
	{
		"time": "<<PRESENCE>>",
		"code": 500,
		"message": "message 1",
		"caller": "<<PRESENCE>>",
		"level": "error",
		"service": {"name": "test"},
		"error": "mock"
	}`)
}

func Test_Error_WriteError(t *testing.T) {
	tests := []struct {
		name   string
		error  Error
		writer func() (LineWriter, fmt.Stringer)
		want   string
	}{
		{
			name: "No Duplicates",
			error: func() Error {
				err := BailFreeze("bail")
				err = WrapFreeze(err, "wrap")
				return Wrap(err).Freeze()
			}(),
			writer: func() (LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "wrap => bail",
		},
		{
			name: "No Duplicates - Tail",
			error: func() Error {
				err := errors.New("errors.New")
				err = WrapFreeze(err, "wrap")
				err = Wrap(err).Freeze()
				return Wrap(err).Message("foo").Freeze()
			}(),
			writer: func() (LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "foo => wrap => errors.New",
		},
		{
			name: "Ensure different messages are written",
			error: func() Error {
				err := BailFreeze("bail")
				err = WrapFreeze(err, "wrap")
				return Wrap(err).Message("wrap 2").Freeze()
			}(),
			writer: func() (LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "wrap 2 => wrap => bail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, buf := tt.writer()
			tt.error.WriteError(writer)
			if got := buf.String(); got != tt.want {
				t.Errorf("Error.WriteError() = %v, want %v", got, tt.want)
			}
		})
	}
}
