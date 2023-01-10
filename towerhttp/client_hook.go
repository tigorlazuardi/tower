package towerhttp

import (
	"golang.org/x/net/context"
	"net/http"
)

type ClientHookContext struct {
	Context      context.Context
	Request      *http.Request
	RequestBody  ClonedBody
	Response     *http.Response
	ResponseBody ClonedBody
	Error        error
}

type ClientHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodySize(req *http.Request, res *http.Response) int
	ExecuteHook(ctx *ClientHookContext)
}

type (
	FilterClientRequest   = func(*http.Request) bool
	FilterClientResponse  = func(*http.Response) bool
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

func (c clientHook) AcceptResponseBodySize(_ *http.Request, res *http.Response) int {
	if c.filterResponse(res) {
		return c.readRespondLimit
	}
	return 0
}

func (c clientHook) ExecuteHook(ctx *ClientHookContext) {
	c.log(ctx)
}
