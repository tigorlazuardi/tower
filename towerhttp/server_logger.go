package towerhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"net/http"
	"strings"
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
	// It is an empty buffer if the response body is nil or if ServerLogger.ReceiveResponseBodyStream returns false.
	ResponseBody ClonedBody
	// Error is the error that has been received when sending the response to client.
	// NOT the error that is passed to Responder.RespondError.
	Error error
	// Caller is where the request is executed.
	Caller tower.Caller
	// Tower is the tower instance that is used by Responder.
	Tower *tower.Tower
}

type ServerLogger interface {
	// ReceiveRequestBody should return value other than 0 if the request body should be cloned for logging.
	// Implementors must not consume the request body at this stage.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire request body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the request body should be cloned.
	ReceiveRequestBody(*http.Request) int
	// ReceiveResponseBodyStream should return other value other than 0 if the response body wants to be cloned.
	//
	// If the user responds with data that is not a stream, the response body will always be sent to ServerLoggerContext.ResponseBody.
	//
	// The returned value is the maximum amount of bytes that is desired to read from the request body.
	//
	// A value of 0 effectively skips the request body cloning.
	// A value of -1 (or any negative value) means that the entire response body should be cloned.
	// A value of n (where n > 0) means that the first n bytes of the response body should be cloned.
	ReceiveResponseBodyStream(responseContentType string, r *http.Request) int
	// Log will be called after the Request-Response trip is done.
	// Whether the log will be printed or not depends on the implementation.
	Log(ctx *ServerLoggerContext)
}

type implServerLogger struct {
	opts *serverLoggerOpts
}

func (i implServerLogger) ReceiveRequestBody(request *http.Request) int {
	if i.opts.requestFilter != nil && !i.opts.requestFilter(request) {
		return 0
	}
	return i.opts.limit
}

func (i implServerLogger) ReceiveResponseBodyStream(responseContentType string, r *http.Request) int {
	if i.opts.responseFilter != nil && !i.opts.responseFilter(responseContentType, r) {
		return 0
	}
	return i.opts.limit
}

func (i implServerLogger) Log(ctx *ServerLoggerContext) {
	url := ctx.Request.Host + ctx.Request.URL.String()
	requestFields := tower.F{
		"method": ctx.Request.Method,
		"url":    url,
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
	responseFields := tower.F{
		"status": ctx.ResponseStatus,
		"header": ctx.ResponseHeader,
	}
	if ctx.ResponseBody.Len() > 0 {
		if strings.Contains(ctx.ResponseHeader.Get("Content-Type"), "application/json") && !ctx.ResponseBody.Truncated() {
			if isJson(ctx.ResponseBody.Bytes()) {
				responseFields["body"] = json.RawMessage(ctx.ResponseBody.CloneBytes())
			} else {
				responseFields["body"] = ctx.ResponseBody.String()
			}
		} else {
			responseFields["body"] = ctx.ResponseBody.String()
		}
	}

	fields := tower.F{
		"request":  requestFields,
		"response": responseFields,
	}
	message := fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.String())
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Message(message).Caller(ctx.Caller).Context(fields).Log(ctx.Context)
		return
	}
	entry := ctx.Tower.NewEntry(message).Caller(ctx.Caller).Context(fields).Log(ctx.Context)
	if i.opts.notify {
		entry.Notify(ctx.Context, i.opts.notifyOption...)
	}
}

// NewServerLogger creates a new built-in implementation of ServerLogger.
func NewServerLogger(opts ...ServerLoggerOption) ServerLogger {
	const defaultLimit = 1024 * 1024 // 1MB
	options := &serverLoggerOpts{
		limit: defaultLimit,
		requestFilter: func(r *http.Request) bool {
			return isHumanReadable(r.Header.Get("Content-Type"))
		},
		responseFilter: func(contentType string, r *http.Request) bool {
			return isHumanReadable(contentType)
		},
	}
	for _, opt := range opts {
		opt.apply(options)
	}
	return implServerLogger{
		opts: options,
	}
}
