package towerhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tigorlazuardi/tower"
)

type ServerLoggerContext struct {
	ServerLoggerMessage
	// Tower is the tower instance that is used by Responder.
	Tower *tower.Tower
}

type ServerLoggerMessage struct {
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
	// Error is the error that has been received when sending the response to client or from Responder.RespondError parameter.
	//
	// If there's no error when sending data to client, Error may be from Responder.RespondError.
	//
	// Otherwise, Error is nil.
	Error error
	// Caller is where the request is executed.
	Caller tower.Caller
	// Level is the log level.
	Level tower.Level
}

type ServerLogger interface {
	// Log will be called after the Request-Response trip is done.
	// Whether the log will be printed or not depends on the implementation.
	Log(ctx *ServerLoggerContext)
}

type implServerLogger struct {
	opts *serverLoggerOpts
}

func isJson(b []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(b, &js) == nil
}

func (i implServerLogger) Log(ctx *ServerLoggerContext) {
	url := ctx.Request.Host + ctx.Request.URL.String()
	requestFields := tower.F{
		"method": ctx.Request.Method,
		"url":    url,
	}
	if len(ctx.Request.Header) > 0 {
		requestFields["headers"] = ctx.Request.Header
	}
	if ctx.RequestBody.Len() > 0 {
		if ctx.RequestBody.Truncated() {
			requestFields["body"] = fmt.Sprintf("%s (truncated)", ctx.RequestBody.String())
		} else if strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
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
	}
	if len(ctx.ResponseHeader) > 0 {
		responseFields["headers"] = ctx.ResponseHeader
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
		_ = ctx.Tower.Wrap(ctx.Error).Level(ctx.Level).Code(ctx.ResponseStatus).Message(message).Caller(ctx.Caller).Context(fields).Log(ctx.Context)
		return
	}
	entry := ctx.Tower.NewEntry(message).Level(ctx.Level).Code(ctx.ResponseStatus).Caller(ctx.Caller).Context(fields).Log(ctx.Context)
	if i.opts.notify {
		entry.Notify(ctx.Context, i.opts.notifyOption...)
	}
}

func isHumanReadable(contentType string) bool {
	return contentType == "" ||
		strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml")
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
