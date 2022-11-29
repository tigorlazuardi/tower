package towerhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/internal/pool"
	"io"
	"net/http"
	"strings"
)

var clientBodyPool = pool.New(func() *bytes.Buffer { return &bytes.Buffer{} })

type clientBodyCloner struct {
	io.ReadCloser
	clone    *bytes.Buffer
	limit    int
	callback func()
}

func (c clientBodyCloner) Close() error {
	clientBodyPool.Put(c.clone)
	if c.callback != nil {
		c.callback()
	}
	return c.ReadCloser.Close()
}

func (c clientBodyCloner) Bytes() []byte {
	return c.clone.Bytes()
}

func (c clientBodyCloner) CloneBytes() []byte {
	s := make([]byte, c.clone.Len())
	copy(s, c.clone.Bytes())
	return s
}

func (c clientBodyCloner) String() string {
	return c.clone.String()
}

func (c clientBodyCloner) Len() int {
	return c.clone.Len()
}

func (c clientBodyCloner) Buffer() io.Reader {
	return c.clone
}

func (c clientBodyCloner) Truncated() bool {
	return c.limit > 0 && c.clone.Len() >= c.limit
}

func (c *clientBodyCloner) Read(p []byte) (n int, err error) {
	n, err = c.Read(p)
	if c.limit > 0 && c.clone.Len() >= c.limit {
		return n, err
	}
	if n > 0 {
		c.clone.Write(p[:n])
	}
	return n, err
}

func wrapClientBodyCloner(r io.Reader, limit int, callback func()) *clientBodyCloner {
	cl := clientBodyPool.Get()
	cl.Reset()
	var rc io.ReadCloser
	if c, ok := r.(io.ReadCloser); ok {
		rc = c
	} else {
		rc = io.NopCloser(r)
	}
	return &clientBodyCloner{
		ReadCloser: rc,
		clone:      cl,
		limit:      limit,
		callback:   callback,
	}
}

type ClonedBody interface {
	// Bytes returns the bytes of the body. The returned byte slice are only valid until next client request.
	//
	// Use CloneBytes to get a copy of the bytes.
	Bytes() []byte
	// CloneBytes returns a copy of the bytes of the body.
	CloneBytes() []byte
	// String returns the body as a string.
	String() string
	// Len returns the number of bytes in the body.
	Len() int
	// Buffer returns a reader that is used to store the body.
	Buffer() io.Reader
	// Truncated returns true if the body was truncated.
	Truncated() bool
}

type noopCloneBody struct{}

func (n noopCloneBody) Bytes() []byte {
	return []byte{}
}

func (n noopCloneBody) CloneBytes() []byte {
	return []byte{}
}

func (n noopCloneBody) String() string {
	return ""
}

func (n noopCloneBody) Len() int {
	return 0
}

func (n noopCloneBody) Buffer() io.Reader {
	return &bytes.Buffer{}
}

func (n noopCloneBody) Truncated() bool {
	return false
}

type ClientRequestContext struct {
	// Context is the context used for the request.
	Context context.Context
	// Request is the request that has been sent. The request body is more than likely have been consumed. It's
	// not advised to consume the body again.
	Request *http.Request
	// RequestBody is a clone of request body that has been sent.
	// It is an empty buffer if the request body is nil or if ClientLogger.ReceiveRequestBody returns false.
	RequestBody ClonedBody
	// Response is the response that has been received. The response body have been consumed. It's pointless to consume
	// the body.
	//
	// Response can be nil if the request failed.
	Response *http.Response
	// ResponseBody is a clone of response body that has been received.
	// It is an empty buffer if the response body is nil or if ClientLogger.ReceiveResponseBody returns false.
	//
	// ResponseBody is always empty (not nil) if Response is nil.
	ResponseBody ClonedBody
	// Error is the error that has been received when sending the request.
	Error error
	// Caller is where the request is executed.
	Caller tower.Caller
}

type ClientLogger interface {
	// ReceiveRequestBody should return true if the request body should be cloned for logging.
	// Implementors must not consume the request body at this stage.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire request body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the request body should be cloned.
	ReceiveRequestBody(*http.Request) int
	// ReceiveResponseBody should return true if the request body should be cloned for logging.
	// Implementors must not consume the response body at this stage.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire request body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the request body should be cloned.
	ReceiveResponseBody(*http.Request, *http.Response) int
	// Log will be called after the Request-Response trip is done.
	// Whether the log will be printed or not depends on the implementation.
	Log(ctx *ClientRequestContext)
}

type towerClientLogger struct {
	t *tower.Tower
}

func isHumanReadable(contentType string) bool {
	return strings.HasPrefix(contentType, "text/") ||
		strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/xml") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded")
}

func isJson(b []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(b, &js) == nil
}

func (t towerClientLogger) ReceiveRequestBody(request *http.Request) int {
	contentType := request.Header.Get("Content-Type")
	if isHumanReadable(contentType) {
		return 1024 * 1024
	}
	return 0
}

func (t towerClientLogger) ReceiveResponseBody(_ *http.Request, response *http.Response) int {
	contentType := response.Header.Get("Content-Type")
	if isHumanReadable(contentType) {
		return 1024 * 1024
	}
	return 0
}

func (t towerClientLogger) Log(ctx *ClientRequestContext) {
	requestFields := tower.F{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
		"header": ctx.Request.Header,
	}
	if ctx.RequestBody.Len() > 0 {
		if strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") && !ctx.RequestBody.Truncated() {
			if isJson(ctx.RequestBody.Bytes()) {
				requestFields["body"] = json.RawMessage(ctx.RequestBody.CloneBytes())
			} else {
				requestFields["body"] = ctx.RequestBody.String()
			}
		} else {
			requestFields["body"] = ctx.RequestBody.String()
		}
	}
	var statusCode = http.StatusInternalServerError

	fields := tower.F{
		"request": requestFields,
	}

	if ctx.Response != nil {
		statusCode = ctx.Response.StatusCode
		responseFields := tower.F{
			"status":   ctx.Response.StatusCode,
			"header":   ctx.Response.Header,
			"protocol": ctx.Response.Proto,
		}
		if !ctx.ResponseBody.Truncated() && ctx.ResponseBody.Len() > 0 {
			if strings.Contains(ctx.Response.Header.Get("Content-Type"), "application/json") {
				responseFields["body"] = json.RawMessage(ctx.ResponseBody.CloneBytes())
			} else if isHumanReadable(ctx.Response.Header.Get("Content-Type")) {
				responseFields["body"] = ctx.ResponseBody.String()
			} else {
				responseFields["body"] = "(binary)"
			}
		}
		fields["response"] = responseFields
	}

	message := fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.String())
	if ctx.Error != nil {
		_ = t.t.
			Wrap(ctx.Error).
			Code(statusCode).
			Message(message).
			Context(fields).
			Caller(ctx.Caller).
			Log(ctx.Context)
		return
	}
	t.t.NewEntry(message).Code(statusCode).Context(fields).Caller(ctx.Caller).Log(ctx.Context)
}

func NewTowerClientLogger(t *tower.Tower) ClientLogger {
	return &towerClientLogger{t: t}
}

type NoopClientLogger struct{}

func (n NoopClientLogger) ReceiveRequestBody(*http.Request) int                  { return 0 }
func (n NoopClientLogger) ReceiveResponseBody(*http.Request, *http.Response) int { return 0 }
func (n NoopClientLogger) Log(*ClientRequestContext)                             {}
