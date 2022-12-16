package towerzap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type mockErrorStruct struct {
	Message string `json:"message,omitempty"`
}

func (m mockErrorStruct) Error() string {
	return m.Message
}

type mockErrorFunc func()

func (m mockErrorFunc) Error() string {
	return "ss"
}

func newTestLogger() (*zap.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		SkipLineEnding: false,
		LineEnding:     "\n",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}), zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)
	return logger, buf
}

func newTower() *tower.Tower {
	return tower.NewTower(tower.Service{
		Name:        "test-towerzap",
		Environment: "testing",
		Repository:  "",
		Branch:      "",
		Type:        "test",
		Version:     "v0.1.0",
	})
}

func prettyPrintJson(buf *bytes.Buffer) {
	if buf.Len() == 0 {
		fmt.Println("Empty buffer")
		return
	}
	out := &bytes.Buffer{}
	b := buf.Bytes()
	err := json.Indent(out, b, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}

func TestLogger_Log(t *testing.T) {
	type args struct {
		ctx   context.Context
		entry tower.Entry
	}
	tests := []struct {
		name          string
		args          args
		traceCapturer TraceCapturer
		test          func(t *testing.T, buf *bytes.Buffer)
	}{
		{
			name: "expected - minimal",
			args: args{
				ctx:   context.Background(),
				entry: newTower().NewEntry("foo").Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "info",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"caller": "<<PRESENCE>>"
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - single context",
			args: args{
				ctx:   context.Background(),
				entry: newTower().NewEntry("foo").Context(tower.F{"buzz": "light-year"}).Freeze(),
			},
			traceCapturer: TraceCapturerFunc(func(ctx context.Context) []zap.Field {
				return []zap.Field{
					zap.String("trace", "captured"),
				}
			}),
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "info",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"trace": "captured",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"caller": "<<PRESENCE>>",
					"context": {
						"buzz": "light-year"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - multiple context",
			args: args{
				ctx: context.Background(),
				entry: newTower().NewEntry("foo").Key("cramp").Context(
					tower.F{"buzz": "light-year"},
					tower.F{"fizz": "buzz", "will": tower.F{"buzz": "fizz"}},
					12345,
					zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
						oe.AddString("zap", "core")
						return nil
					}),
					zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
						ae.AppendBool(true)
						return nil
					}),
					func() {},
					errors.New("foo"),
				).Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "info",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"key": "cramp",
					"caller": "<<PRESENCE>>",
					"context": [
						{
							"buzz": "light-year"
						},
						{
							"fizz": "buzz",
							"will": {
								"buzz": "fizz"
							}
						},
						12345,
						{
							"zap": "core"
						},
						[
							true
						],
						"json: unsupported type: func()",
						{"error": {"summary": "foo"}}
					]
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - entry in context",
			args: args{
				ctx: context.Background(),
				entry: func() tower.Entry {
					tow := newTower()
					return tow.NewEntry("foo").Key("fizz").Code(200).Context(
						tower.F{"entry": tow.NewEntry("bar").Context(func() {}).Freeze()},
					).Freeze()
				}(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "info",
					"message": "foo",
					"code": 200,
					"key": "fizz",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"caller": "<<PRESENCE>>",
					"context": {
						"entry": {
							"level": "info", 
							"message": 
							"bar", 
							"caller": "<<PRESENCE>>",
							"context": "<<PRESENCE>>"
						}
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := newTestLogger()
			l := NewLogger(logger)
			if tt.traceCapturer != nil {
				l.SetTraceCapturer(tt.traceCapturer)
			}
			l.Log(tt.args.ctx, tt.args.entry)
			tt.test(t, buf)
			if t.Failed() {
				prettyPrintJson(buf)
			}
		})
	}
}

func TestLogger_LogError(t *testing.T) {
	type args struct {
		ctx context.Context
		err tower.Error
	}
	tests := []struct {
		name          string
		args          args
		traceCapturer TraceCapturer
		test          func(t *testing.T, buf *bytes.Buffer)
	}{
		{
			name: "origin error is nil",
			args: args{
				ctx: context.Background(),
				err: newTower().Wrap(nil).Message("wack").Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "wack",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"code": 500,
					"caller": "<<PRESENCE>>",
					"error": {
						"summary": "<nil>"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - minimal",
			args: args{
				ctx: context.Background(),
				err: newTower().Bail("foo").Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"code": 500,
					"caller": "<<PRESENCE>>",
					"error": {
						"summary": "foo"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - non marshalable error",
			args: args{
				ctx: context.Background(),
				err: newTower().Wrap(mockErrorFunc(func() {})).Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "ss",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"code": 500,
					"caller": "<<PRESENCE>>",
					"error": {
						"summary": "ss",
						"details": "json: unsupported type: towerzap.mockErrorFunc"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - one context",
			args: args{
				ctx: context.Background(),
				err: newTower().Bail("foo").Context(zapcore.ObjectMarshalerFunc(func(encoder zapcore.ObjectEncoder) error {
					encoder.AddString("foo", "bar")
					return nil
				})).Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"code": 500,
					"caller": "<<PRESENCE>>",
					"context": {
						"foo": "bar"
					},
					"error": {
						"summary": "foo"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "expected - multiple context",
			args: args{
				ctx: context.Background(),
				err: newTower().Bail("foo").Context(
					zapcore.ObjectMarshalerFunc(func(encoder zapcore.ObjectEncoder) error {
						encoder.AddString("foo", "bar")
						return nil
					}),
					12345,
				).Freeze(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"code": 500,
					"caller": "<<PRESENCE>>",
					"context": [{"foo": "bar"}, 12345],
					"error": {
						"summary": "foo"
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
			},
		},
		{
			name: "nested tower.Error",
			args: args{
				ctx: context.Background(),
				err: func() tower.Error {
					tow := newTower()
					inner := tow.Wrap(mockErrorStruct{"bar"}).Key("buzz").Context(54321, 33333).Freeze()
					return tow.Wrap(inner).Key("fizz").Message("foo").Context(
						zapcore.ArrayMarshalerFunc(func(encoder zapcore.ArrayEncoder) error {
							encoder.AppendInt(12345)
							return nil
						}),
					).Freeze()
				}(),
			},
			traceCapturer: nil,
			test: func(t *testing.T, buf *bytes.Buffer) {
				j := jsonassert.New(t)
				got := buf.String()
				want := `
				{
					"level": "error",
					"message": "foo",
					"time": "<<PRESENCE>>",
					"service": {
						"name": "test-towerzap",
						"type": "test",
						"environment": "testing",
						"version": "v0.1.0"
					},
					"key": "fizz",
					"code": 500,
					"caller": "<<PRESENCE>>",
					"context": [
						12345
					],
					"error": {
						"code": 500,
						"message": "bar",
						"caller": "<<PRESENCE>>",
						"context": [54321, 33333],
						"key": "buzz",
						"error": {
							"summary": "bar",
							"details": {"message": "bar"}
						}
					}
				}`
				j.Assertf(got, want)
				if !strings.Contains(got, "towerzap_test.go:") {
					t.Error("want caller to be on this file")
				}
				if strings.Count(got, "towerzap_test.go:") != 2 {
					t.Error("want caller to be on this file twice")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := newTestLogger()
			l := NewLogger(logger)
			if tt.traceCapturer != nil {
				l.SetTraceCapturer(tt.traceCapturer)
			}
			l.LogError(tt.args.ctx, tt.args.err)
			tt.test(t, buf)
			if t.Failed() {
				prettyPrintJson(buf)
			}
		})
	}
}
