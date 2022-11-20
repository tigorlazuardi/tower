package towerhttp

import (
	"bytes"
	"context"
	"net/http"
)

type RequestContext interface {
	// Context Returns the context for the request.
	Context() context.Context
	// Request Returns the request. Note that the body is already consumed and cannot be consumed again.
	Request() *http.Request
	// RequestBody returns the cloned request body, body is only cloned if Logger.ReceiveRequestBody returns true.
	//
	// RequestBody never returns nil, but will return an empty bytes.Buffer if Logger.ReceiveRequestBody is false, or
	// if the request body is empty.
	//
	// The buffer is guaranteed to last only inside the Logger.Log scope before it is sent back to the pool,
	// The underlying array will be rewritten with other values upon new Request.
	// If you need to keep the body for elsewhere, you must copy the bytes inside this buffer.
	RequestBody() *bytes.Buffer
	// Response Returns the response. Note that the body is already consumed and cannot be consumed again.
	Response() *http.Response
	// ResponseBody returns the cloned response body, body is only cloned if Logger.ReceiveResponseBody returns true.
	//
	// ResponseBody never returns nil, but will return an empty bytes.Buffer if Logger.ReceiveResponseBody is false or
	// if the response body is empty or Client.Do fails.
	//
	// The buffer is guaranteed to last only inside the Logger.Log scope before it is sent back to the pool,
	// The underlying array will be rewritten with other values upon new Request.
	// If you need to keep the body for elsewhere, you must copy the bytes inside this buffer.
	ResponseBody() *bytes.Buffer
	// Error returns the error if any error happens in Client.Do request.
	Error() error
}

type Logger interface {
	// ReceiveRequestBody should return true if the request body should be cloned for logging.
	// Implementors must not consume the request body at this stage.
	ReceiveRequestBody(*http.Request) bool
	// ReceiveResponseBody should return true if the request body should be logged.
	// Implementors must not consume the response body at this stage.
	ReceiveResponseBody(*http.Response) bool
	// Log will be called after the Request-Response trip is done.
	// Whether the log will be printed or not depends on the implementation.
	Log(ctx RequestContext)
}
