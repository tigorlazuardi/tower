package towerhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tigorlazuardi/tower"
)

type RoundTripHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodySize(req *http.Request, res *http.Response) int
	ExecuteHook(ctx *RoundTripContext)
}

func NewRoundTripHook(opts ...RoundTripHookOption) RoundTripHook {
	hook := &roundTripHook{}
	opts = append(defaultRoundTriptHookOpts(), opts...)
	for _, v := range opts {
		v.apply(hook)
	}
	return hook
}

type (
	RoundTripFilterRequest   = func(*http.Request) bool
	RoundTripFilterResponse  = func(*http.Request, *http.Response) bool
	RoundTripExecuteHookFunc = func(*RoundTripContext)
)

type roundTripHook struct {
	readRespondLimit int
	readRequestLimit int
	filterRequest    RoundTripFilterRequest
	filterResponse   RoundTripFilterResponse
	log              RoundTripExecuteHookFunc
}

func (rth *roundTripHook) AcceptRequestBodySize(r *http.Request) int {
	if rth.filterRequest(r) {
		return rth.readRequestLimit
	}
	return 0
}

func (rth *roundTripHook) AcceptResponseBodySize(req *http.Request, res *http.Response) int {
	if rth.filterResponse(req, res) {
		return rth.readRespondLimit
	}
	return 0
}

func (rth *roundTripHook) ExecuteHook(ctx *RoundTripContext) {
	rth.log(ctx)
}

type RoundTripHookOption interface {
	apply(*roundTripHook)
}

type (
	RoundTripHookOptionBuilder []RoundTripHookOption
	RoundTripHookOptionFunc    func(*roundTripHook)
)

func (rt RoundTripHookOptionBuilder) apply(hook *roundTripHook) {
	for _, opt := range rt {
		opt.apply(hook)
	}
}

func (rt RoundTripHookOptionFunc) apply(hook *roundTripHook) {
	rt(hook)
}

func (r RoundTripHookOptionBuilder) ReadRequestBodyLimit(limit int) RoundTripHookOptionBuilder {
	return append(r, RoundTripHookOptionFunc(func(hook *roundTripHook) {
		hook.readRequestLimit = limit
	}))
}

func (r RoundTripHookOptionBuilder) ReadResponseBodyLimit(limit int) RoundTripHookOptionBuilder {
	return append(r, RoundTripHookOptionFunc(func(hook *roundTripHook) {
		hook.readRespondLimit = limit
	}))
}

func (r RoundTripHookOptionBuilder) FilterRequest(filter RoundTripFilterRequest) RoundTripHookOptionBuilder {
	return append(r, RoundTripHookOptionFunc(func(hook *roundTripHook) {
		hook.filterRequest = filter
	}))
}

func (r RoundTripHookOptionBuilder) FilterResponse(filter RoundTripFilterResponse) RoundTripHookOptionBuilder {
	return append(r, RoundTripHookOptionFunc(func(hook *roundTripHook) {
		hook.filterResponse = filter
	}))
}

func (r RoundTripHookOptionBuilder) Log(log RoundTripExecuteHookFunc) RoundTripHookOptionBuilder {
	return append(r, RoundTripHookOptionFunc(func(hook *roundTripHook) {
		hook.log = log
	}))
}

func defaultRoundTriptHookOpts() RoundTripHookOptionBuilder {
	return Option.RoundTripHook().
		ReadRequestBodyLimit(1024 * 1024).
		ReadResponseBodyLimit(1024 * 1024).
		FilterRequest(func(r *http.Request) bool { return isHumanReadable(r.Header.Get("Content-Type")) }).
		FilterResponse(func(_ *http.Request, res *http.Response) bool { return isHumanReadable(res.Header.Get("Content-Type")) }).
		Log(defaultRounTripLogFunc)
}

func defaultRounTripLogFunc(ctx *RoundTripContext) {
	reqCtx := ctx.Request.Context()
	fields := tower.F{}
	if ctx.Response != nil {
		fields = buildClientResponseFields(fields, ctx.Response, ctx.ResponseBody)
	}
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Context(fields).Log(reqCtx)
	} else if ctx.Response != nil && ctx.Response.StatusCode >= 400 {
		_ = ctx.Tower.Bail("error: %s %s. %s", ctx.Request.Method, ctx.Request.URL, ctx.Response.Status).
			Context(fields).
			Log(reqCtx)
	} else {
		ctx.Tower.
			NewEntry("success: %s %s", ctx.Request.Method, ctx.Request.URL).
			Context(fields).
			Log(reqCtx)
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
