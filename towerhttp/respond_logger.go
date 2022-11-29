package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type ServerLoggerContext struct {
	// Context is the context used for the request.
	Context context.Context
	// Request is the request that has been sent. The request body is more than likely have been consumed. It's
	// not advised to consume the body again.
	Request *http.Request
	// RequestBody is a clone of request body that has been sent.
	// It is an empty buffer if the request body is nil or if ServerLogger.ReceiveRequestBody returns false.
	RequestBody ClonedBody
	// ResponseStatus response status code.
	ResponseStatus int
	// ResponseHeader is the response that has been received. The response body have been consumed. It's pointless to consume
	// the body.
	ResponseHeader http.Header
	// ResponseBody is a clone of response body that has been received.
	// It is an empty buffer if the response body is nil or if ServerLogger.ReceiveResponseBody returns false.
	ResponseBody ClonedBody
	// Error is the error that has been received when sending the request.
	Error error
	// Caller is where the request is executed.
	Caller tower.Caller
}

type ServerLogger interface {
	// ReceiveRequestBody should return true if the request body should be cloned for logging.
	// Implementors must not consume the request body at this stage.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire request body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the request body should be cloned.
	ReceiveRequestBody(*http.Request) int
	// ReceiveResponseBody should return true if the response body should be cloned for logging.
	// Implementors must not consume the request body at this stage.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire request body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the request body should be cloned.
	ReceiveResponseBody(responseContentType string, r *http.Request) int
	// Log will be called after the Request-Response trip is done.
	// Whether the log will be printed or not depends on the implementation.
	Log(ctx *ServerLoggerContext)
}
