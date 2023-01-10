package towerhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tigorlazuardi/tower"
)

type ClientHookContext struct {
	Context context.Context
	// Request may be nil if towerhttp have no way to reach the http.Request instance.
	Request     *http.Request
	RequestBody ClonedBody
	// Response may be nil if
	Response     *http.Response
	ResponseBody ClonedBody
	Error        error
	Tower        *tower.Tower
}

type ClientHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodySize(req *http.Request, res *http.Response) int
	ExecuteHook(ctx *ClientHookContext)
}

type (
	FilterClientRequest   = func(*http.Request) bool
	FilterClientResponse  = func(*http.Request, *http.Response) bool
	ClientExecuteHookFunc = func(*ClientHookContext)
)

type clientHook struct {
	readRequestLimit int
	readRespondLimit int
	filterRequest    FilterRequest
	filterResponse   FilterClientResponse
	log              ClientExecuteHookFunc
}

func (c clientHook) AcceptRequestBodySize(r *http.Request) int {
	if c.filterRequest(r) {
		return c.readRequestLimit
	}
	return 0
}

func (c clientHook) AcceptResponseBodySize(req *http.Request, res *http.Response) int {
	if c.filterResponse(req, res) {
		return c.readRespondLimit
	}
	return 0
}

func (c clientHook) ExecuteHook(ctx *ClientHookContext) {
	c.log(ctx)
}

func NewClientLoggerHook(opts ...ClientHookOption) ClientHook {
	c := &clientHook{}
	opts = append(defaultClientLoggerOptions(), opts...)
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func defaultClientLoggerOptions() ClientHookOptionBuilder {
	return Option.ClientHook().
		ReadRequestBodyLimit(1024 * 1024).
		ReadResponseBodyLimit(1024 * 1024).
		FilterRequest(func(r *http.Request) bool {
			return isHumanReadable(r.Header.Get("Content-Type"))
		}).
		FilterResponse(func(_ *http.Request, res *http.Response) bool {
			return isHumanReadable(res.Header.Get("Content-Type"))
		}).
		Log(defaultClientLogFunc)
}

func defaultClientLogFunc(ctx *ClientHookContext) {
	// Towerhttp may fail to fetch request data from Response instance.
	if ctx.Request == nil {
		return
	}
	fields := tower.F{}
	if ctx.Response != nil {
		fields = buildClientResponseFields(fields, ctx.Response, ctx.ResponseBody)
	}
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Context(fields).Log(ctx.Context)
	} else if ctx.Response != nil && ctx.Response.StatusCode >= 400 {
		_ = ctx.Tower.Bail("error: %s %s. %s", ctx.Request.Method, ctx.Request.URL, ctx.Response.Status).
			Context(fields).
			Log(ctx.Context)
	} else {
		ctx.Tower.
			NewEntry("success: %s %s", ctx.Request.Method, ctx.Request.URL).
			Context(fields).
			Log(ctx.Context)
	}
}

func buildClientRequestFields(f tower.Fields, req *http.Request, body ClonedBody) tower.Fields {
	fields := tower.F{
		"method": req.Method,
		"url":    req.URL.String(),
		"header": req.Header,
	}

	if body.Len() > 0 {
		contentType := req.Header.Get("Content-Type")
		switch {
		case body.Truncated():
			fields["body"] = fmt.Sprintf("%s (truncated)", body.String())
		case strings.Contains(contentType, "application/json") && isJson(body.Bytes()):
			fields["body"] = json.RawMessage(body.CloneBytes())
		case contentType == "" && isJsonLite(body.Bytes()) && isJson(body.Bytes()):
			fields["body"] = json.RawMessage(body.CloneBytes())
		default:
			fields["body"] = body.String()
		}
	}

	f["request"] = fields
	return f
}

func buildClientResponseFields(f tower.Fields, res *http.Response, body ClonedBody) tower.Fields {
	fields := tower.F{
		"status": res.Status,
		"header": res.Header,
	}
	if body.Len() > 0 {
		contentType := res.Header.Get("Content-Type")
		switch {
		case body.Truncated():
			fields["body"] = fmt.Sprintf("%s (truncated)", body.String())
		case strings.Contains(contentType, "application/json") && isJson(body.Bytes()):
			fields["body"] = json.RawMessage(body.CloneBytes())
		case contentType == "" && isJsonLite(body.Bytes()) && isJson(body.Bytes()):
			fields["body"] = json.RawMessage(body.CloneBytes())
		default:
			fields["body"] = body.String()
		}
	}
	f["response"] = fields
	return f
}
