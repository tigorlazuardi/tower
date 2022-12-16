package towerzap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
		Version:     "",
	})
}

func prettyPrintJson(buf *bytes.Buffer) {
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
						"environment": "testing"
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
						"environment": "testing"
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
						"environment": "testing"
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
						]
					]
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
	type fields struct {
		Logger *zap.Logger
		tracer TraceCapturer
	}
	type args struct {
		ctx context.Context
		err tower.Error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Logger{
				Logger: tt.fields.Logger,
				tracer: tt.fields.tracer,
			}
			l.LogError(tt.args.ctx, tt.args.err)
		})
	}
}
