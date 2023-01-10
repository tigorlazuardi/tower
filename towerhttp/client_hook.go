package towerhttp

import (
	"encoding/json"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

type ClientHookContext struct {
	Context      context.Context
	Request      *http.Request
	RequestBody  ClonedBody
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

func defaultClientLoggerOptions() []ClientHookOption {
	return []ClientHookOption{
		ClientHookOptionFunc(func(hook *clientHook) {
			hook.readRequestLimit = 1024
			hook.readRespondLimit = 1024
			hook.filterRequest = func(r *http.Request) bool {
				return isHumanReadable(r.Header.Get("Content-Type"))
			}
			hook.filterResponse = func(_ *http.Request, res *http.Response) bool {
				return isHumanReadable(res.Header.Get("Content-Type"))
			}
			hook.log = func(ctx *ClientHookContext) {
				fields := tower.F{}
				fields = buildClientRequestFields(fields, ctx.Request, ctx.RequestBody)
				if ctx.Error != nil {
					_ = ctx.Tower.Wrap(ctx.Error).Context(fields).Log(ctx.Context)
				} else {
					ctx.Tower.
						NewEntry("%s %s", ctx.Request.Method, ctx.Request.URL.String()).
						Context(fields).
						Log(ctx.Context)
				}
			}
		}),
	}
}

func buildClientRequestFields(f tower.Fields, req *http.Request, body ClonedBody) tower.Fields {
	fields := tower.F{
		"method": req.Method,
		"url":    req.Proto + "://" + req.Host + req.URL.String(),
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
