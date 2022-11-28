package tower

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// ErrorNode is the implementation of the Error interface.
type ErrorNode struct {
	inner *errorBuilder
	flag  marshalFlag
}

// sorted keys are rather important for human reads. Especially the Context and Error should always be at the last marshaled keys.
// as they contain the most amount of data and information, and thus shadows other values at a glance.
//
// arguably this is simpler to be done than implementing json.Marshaler interface and doing it manually, key by key
// without resorting to other libraries.
type implErrorJsonMarshaler struct {
	Time    string `json:"time,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Caller  Caller `json:"caller,omitempty"`
	Key     string `json:"key,omitempty"`
	Level   string `json:"level,omitempty"`
	Context any    `json:"context,omitempty"`
	Error   error  `json:"error,omitempty"`
}

type marshalFlag uint8

func (m marshalFlag) Has(f marshalFlag) bool {
	return m&f == f
}

func (m *marshalFlag) Set(f marshalFlag) {
	*m |= f
}

const (
	marshalAll      marshalFlag = 0
	marshalSkipCode marshalFlag = 1 << iota
	marshalSkipMessage
	marshalSkipLevel
	marshalSkipCaller
	marshalSkipContext
	marshalSkipTime
	marshalSkipAll = marshalSkipCode + marshalSkipMessage + marshalSkipLevel + marshalSkipTime + marshalSkipContext + marshalSkipCaller
)

type CodeBlockJSONMarshaler interface {
	CodeBlockJSON() ([]byte, error)
}

type cbJson struct {
	inner error
}

func (c cbJson) Error() string {
	return c.inner.Error()
}

func (c cbJson) CodeBlockJSON() ([]byte, error) {
	return c.MarshalJSON()
}

func (c cbJson) MarshalJSON() ([]byte, error) {
	if cb, ok := c.inner.(CodeBlockJSONMarshaler); ok {
		return cb.CodeBlockJSON()
	}
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(richJsonError{c.inner})
	return b.Bytes(), err
}

func (e *ErrorNode) compareAndUpdateFlags(other *ErrorNode) {
	if e.Code() == other.Code() {
		other.flag.Set(marshalSkipCode)
	}
	if e.Message() == other.Message() {
		other.flag.Set(marshalSkipMessage)
	}
	if e.Level() == other.Level() {
		other.flag.Set(marshalSkipLevel)
	}
	if len(other.Context()) == 0 {
		other.flag.Set(marshalSkipContext)
	}
	if e.Time().Sub(other.Time()) < time.Second {
		other.flag.Set(marshalSkipTime)
	}
	if other.flag.Has(marshalSkipCode) &&
		other.flag.Has(marshalSkipMessage) &&
		other.flag.Has(marshalSkipLevel) &&
		other.flag.Has(marshalSkipContext) {
		other.flag.Set(marshalSkipCaller)
	}
}

func (e *ErrorNode) createCodeBlockPayload() *implErrorJsonMarshaler {
	ctx := func() any {
		if len(e.inner.context) == 0 {
			return nil
		}
		if len(e.inner.context) == 1 {
			return e.inner.context[0]
		}
		return e.inner.context
	}()
	marshalAble := implErrorJsonMarshaler{
		Time:    e.Time().Format(time.RFC3339),
		Code:    e.Code(),
		Message: e.Message(),
		Caller:  e.Caller(),
		Key:     e.Key(),
		Level:   e.Level().String(),
		Context: ctx,
		Error:   cbJson{e.inner.origin},
	}

	if e.flag.Has(marshalSkipCode) {
		marshalAble.Code = 0
	}
	if e.flag.Has(marshalSkipMessage) {
		marshalAble.Message = ""
	}
	if e.flag.Has(marshalSkipLevel) {
		marshalAble.Level = ""
	}
	if e.flag.Has(marshalSkipTime) {
		marshalAble.Time = ""
	}
	if e.flag.Has(marshalSkipCaller) {
		marshalAble.Caller = nil
	}
	return &marshalAble
}

func (e ErrorNode) CodeBlockJSON() ([]byte, error) {
	if origin, ok := e.inner.origin.(*ErrorNode); ok {
		e.compareAndUpdateFlags(origin)
	}
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	// Check if current ErrorNode needs to be skipped.
	if e.flag.Has(marshalSkipAll) {
		origin := e.inner.origin
		if cbJson, ok := origin.(CodeBlockJSONMarshaler); ok {
			return cbJson.CodeBlockJSON()
		}
		err := enc.Encode(richJsonError{origin})
		return b.Bytes(), err
	}
	err := enc.Encode(e.createCodeBlockPayload())
	return b.Bytes(), err
}

func (e ErrorNode) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	ctx := func() any {
		if len(e.inner.context) == 0 {
			return nil
		}
		if len(e.inner.context) == 1 {
			return e.inner.context[0]
		}
		return e.inner.context
	}()

	err := enc.Encode(implErrorJsonMarshaler{
		Time:    e.Time().Format(time.RFC3339),
		Code:    e.Code(),
		Message: e.Message(),
		Caller:  e.Caller(),
		Key:     e.Key(),
		Level:   e.Level().String(),
		Context: ctx,
		Error:   richJsonError{e.inner.origin},
	})
	return b.Bytes(), err
}

func (e ErrorNode) Error() string {
	s := &strings.Builder{}
	lw := NewLineWriter(s).LineBreak(" => ").Build()
	e.WriteError(lw)
	return s.String()
}

// WriteError Writes the error.Error to the writer instead of being allocated as value.
func (e ErrorNode) WriteError(w LineWriter) {
	w.WriteIndent()
	msg := e.inner.message
	if e.inner.origin == nil {
		// Account for empty string message after wrapping nil error.
		if len(msg) > 0 {
			w.WritePrefix()
			_, _ = w.WriteString(msg)
			w.WriteSuffix()
			w.WriteLineBreak()
		}
		w.WritePrefix()
		_, _ = w.WriteString("[nil]")
		w.WriteSuffix()
		return
	}

	writeInner := func(linebreak bool) {
		if ew, ok := e.inner.origin.(ErrorWriter); ok {
			if linebreak {
				w.WriteLineBreak()
			}
			ew.WriteError(w)
		} else {
			errMsg := e.inner.origin.Error()
			if errMsg != msg {
				w.WriteLineBreak()
				w.WritePrefix()
				_, _ = w.WriteString(errMsg)
				w.WriteSuffix()
			}
		}
	}

	var innerMessage string
	if mh, ok := e.inner.origin.(MessageHint); ok {
		innerMessage = mh.Message()
	}

	// Skip writing duplicate or empty messages.
	if msg == innerMessage || len(msg) == 0 {
		writeInner(false)
		return
	}

	w.WritePrefix()
	_, _ = w.WriteString(msg)
	w.WriteSuffix()
	writeInner(true)
}

// Code Gets the original code of the type.
func (e ErrorNode) Code() int {
	return e.inner.code
}

// HTTPCode Gets HTTP Status Code for the type.
func (e ErrorNode) HTTPCode() int {
	switch {
	case e.inner.code >= 200 && e.inner.code <= 599:
		return e.inner.code
	case e.inner.code > 999:
		code := e.inner.code % 1000
		if code >= 200 && code <= 599 {
			return code
		}
	}
	return 500
}

// Message Gets the Message of the type.
func (e ErrorNode) Message() string {
	return e.inner.message
}

// Caller Gets the caller of this type.
func (e ErrorNode) Caller() Caller {
	return e.inner.caller
}

// Context Gets the context of this type.
func (e ErrorNode) Context() []any {
	return e.inner.context
}

func (e ErrorNode) Level() Level {
	return e.inner.level
}

func (e ErrorNode) Time() time.Time {
	return e.inner.time
}

func (e ErrorNode) Key() string {
	return e.inner.key
}

// Unwrap Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
func (e ErrorNode) Unwrap() error {
	return e.inner.origin
}

// Log this error.
func (e ErrorNode) Log(ctx context.Context) Error {
	e.inner.tower.LogError(ctx, e)
	return e
}

// Notify this error to Messengers.
func (e ErrorNode) Notify(ctx context.Context, opts ...MessageOption) Error {
	e.inner.tower.NotifyError(ctx, e, opts...)
	return e
}

// richJsonError is a special kind of error that tries to prevent information loss when marshaling to json.
type richJsonError struct {
	error
}

func (r richJsonError) MarshalJSON() ([]byte, error) {
	if r.error == nil {
		return []byte("null"), nil
	}
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	// if the error supports json.Marshaler we use it directly.
	// this is because we can assume that the error have special marshaling needs for specific output.
	//
	// E.G. to prevent unnecessary "summary" keys when the origin error is already is a tower.Error type.
	if e, ok := r.error.(json.Marshaler); ok { //nolint
		return e.MarshalJSON()
	}

	err := enc.Encode(r.error)
	if err != nil {
		return b.Bytes(), err
	}

	summary := r.error.Error()
	// 3 because it also includes newline after brackets or quotes.
	if b.Len() == 3 && b.Bytes()[2] == '\n' {
		v := b.Bytes()
		switch {
		case v[0] == '"', v[0] == '{', v[0] == '[':
			b.Reset()
			err := enc.Encode(map[string]string{"summary": summary})
			return b.Bytes(), err
		}
	}

	content := b.String()
	b.Reset()
	err = enc.Encode(map[string]json.RawMessage{
		"value":   json.RawMessage(content),
		"summary": json.RawMessage(strconv.Quote(summary)),
	})
	return b.Bytes(), err
}