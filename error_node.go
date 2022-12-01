package tower

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

const codeBlockIndent = "   "

// ErrorNode is the implementation of the Error interface.
type ErrorNode struct {
	inner *errorBuilder
	prev  *ErrorNode
	next  *ErrorNode
}

// sorted keys are rather important for human reads. Especially the Context and Error should always be at the last marshaled keys.
// as they contain the most amount of data and information, and thus shadows other values at a glance.
//
// arguably this is simpler to be done than implementing json.Marshaler interface and doing it manually, key by key
// without resorting to other libraries.
type implJsonMarshaler struct {
	Time    string   `json:"time,omitempty"`
	Code    int      `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
	Caller  Caller   `json:"caller,omitempty"`
	Key     string   `json:"key,omitempty"`
	Level   string   `json:"level,omitempty"`
	Service *Service `json:"service,omitempty"`
	Context any      `json:"context,omitempty"`
	Error   error    `json:"error,omitempty"`
}

type marshalFlag uint8

func (m marshalFlag) Has(f marshalFlag) bool {
	return m&f == f
}

func (m *marshalFlag) Set(f marshalFlag) {
	*m |= f
}

const (
	marshalSkipCode marshalFlag = 1 << iota
	marshalSkipMessage
	marshalSkipLevel
	marshalSkipCaller
	marshalSkipContext
	marshalSkipTime
	marshalSkipService
	marshalSkipAll = marshalSkipCode +
		marshalSkipMessage +
		marshalSkipLevel +
		marshalSkipTime +
		marshalSkipContext +
		marshalSkipCaller +
		marshalSkipService
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
	if c.inner == nil {
		return []byte("null"), nil
	}
	if cb, ok := c.inner.(CodeBlockJSONMarshaler); ok && cb != nil {
		return cb.CodeBlockJSON()
	}
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", codeBlockIndent)
	err := enc.Encode(richJsonError{c.inner})
	return b.Bytes(), err
}

func (e *ErrorNode) createCodeBlockFlagNode() marshalFlag {
	var m marshalFlag
	if e.prev == nil {
		return m
	}
	originIsNode := e.next != nil
	prev, current := e.prev, e
	if prev.Code() == current.Code() && originIsNode {
		m.Set(marshalSkipCode)
	}
	if prev.Message() == current.Message() && originIsNode {
		m.Set(marshalSkipMessage)
	}
	if prev.Level() == current.Level() && originIsNode {
		m.Set(marshalSkipLevel)
	}
	if len(current.Context()) == 0 {
		m.Set(marshalSkipContext)
	}
	if prev.Time().Sub(current.Time()) < time.Second && originIsNode {
		m.Set(marshalSkipTime)
	}
	if m.Has(marshalSkipCode) &&
		m.Has(marshalSkipMessage) &&
		m.Has(marshalSkipLevel) &&
		m.Has(marshalSkipContext) {
		m.Set(marshalSkipCaller)
	}
	if prev.inner.tower.service == current.inner.tower.service {
		m.Set(marshalSkipService)
	}

	return m
}

func (e *ErrorNode) createCodeBlockFlag(other Error) marshalFlag {
	var m marshalFlag
	if e.Code() == other.Code() {
		m.Set(marshalSkipCode)
	}
	if e.Message() == other.Message() {
		m.Set(marshalSkipMessage)
	}
	if e.Level() == other.Level() {
		m.Set(marshalSkipLevel)
	}
	if len(other.Context()) == 0 {
		m.Set(marshalSkipContext)
	}
	if e.Time().Sub(other.Time()) < time.Second {
		m.Set(marshalSkipTime)
	}
	if m.Has(marshalSkipCode) &&
		m.Has(marshalSkipMessage) &&
		m.Has(marshalSkipLevel) &&
		m.Has(marshalSkipContext) {
		m.Set(marshalSkipCaller)
	}
	return m
}

func (e *ErrorNode) createCodeBlockPayload(m marshalFlag) *implJsonMarshaler {
	ctx := func() any {
		if len(e.inner.context) == 0 {
			return nil
		}
		if len(e.inner.context) == 1 {
			return e.inner.context[0]
		}
		return e.inner.context
	}()
	var next error
	if e.next != nil {
		next = e.next
	} else {
		next = e.inner.origin
	}
	marshalAble := implJsonMarshaler{
		Time:    e.Time().Format(time.RFC3339),
		Code:    e.Code(),
		Message: e.Message(),
		Caller:  e.Caller(),
		Key:     e.Key(),
		Level:   e.Level().String(),
		Context: ctx,
		Error:   cbJson{next},
		Service: &e.inner.tower.service,
	}

	if m.Has(marshalSkipCode) {
		marshalAble.Code = 0
	}
	if m.Has(marshalSkipMessage) {
		marshalAble.Message = ""
	}
	if m.Has(marshalSkipLevel) {
		marshalAble.Level = ""
	}
	if m.Has(marshalSkipTime) {
		marshalAble.Time = ""
	}
	if m.Has(marshalSkipCaller) {
		marshalAble.Caller = nil
	}
	if m.Has(marshalSkipService) {
		marshalAble.Service = nil
	}
	return &marshalAble
}

func (e *ErrorNode) CodeBlockJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	var m marshalFlag
	if e.prev != nil {
		m = e.createCodeBlockFlagNode()
	} else if origin, ok := e.inner.origin.(Error); ok {
		m = e.createCodeBlockFlag(origin)
	}
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", codeBlockIndent)
	// Check if current ErrorNode needs to be skipped.
	if m.Has(marshalSkipAll) {
		origin := e.inner.origin
		if cbJson, ok := origin.(CodeBlockJSONMarshaler); ok && cbJson != nil {
			return cbJson.CodeBlockJSON()
		}
		err := enc.Encode(richJsonError{origin})
		return b.Bytes(), err
	}
	err := enc.Encode(e.createCodeBlockPayload(m))
	return bytes.TrimSpace(b.Bytes()), err
}

func (e *ErrorNode) MarshalJSON() ([]byte, error) {
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

	err := enc.Encode(implJsonMarshaler{
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

func (e *ErrorNode) Error() string {
	s := &strings.Builder{}
	lw := NewLineWriter(s).LineBreak(" => ").Build()
	e.WriteError(lw)
	return s.String()
}

// WriteError Writes the error.Error to the writer instead of being allocated as value.
func (e *ErrorNode) WriteError(w LineWriter) {
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
func (e *ErrorNode) Code() int {
	return e.inner.code
}

// HTTPCode Gets HTTP Status Code for the type.
func (e *ErrorNode) HTTPCode() int {
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
func (e *ErrorNode) Message() string {
	return e.inner.message
}

// Caller Gets the caller of this type.
func (e *ErrorNode) Caller() Caller {
	return e.inner.caller
}

// Context Gets the context of this type.
func (e *ErrorNode) Context() []any {
	return e.inner.context
}

func (e *ErrorNode) Level() Level {
	return e.inner.level
}

func (e *ErrorNode) Time() time.Time {
	return e.inner.time
}

func (e *ErrorNode) Key() string {
	return e.inner.key
}

func (e *ErrorNode) Service() Service {
	return e.inner.tower.service
}

// Unwrap Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
func (e *ErrorNode) Unwrap() error {
	return e.inner.origin
}

// Log this error.
func (e *ErrorNode) Log(ctx context.Context) Error {
	e.inner.tower.LogError(ctx, e)
	return e
}

// Notify this error to Messengers.
func (e *ErrorNode) Notify(ctx context.Context, opts ...MessageOption) Error {
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
