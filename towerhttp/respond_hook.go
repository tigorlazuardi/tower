package towerhttp

import (
	"net/http"

	"github.com/tigorlazuardi/tower"
)

type baseHook struct {
	Context        *RespondContext
	Request        *http.Request
	RequestBody    ClonedBody
	ResponseStatus int
	ResponseHeader http.Header
	Tower          *tower.Tower
	Error          error
}

type RespondBody struct {
	PreEncoded     any
	PostEncoded    []byte
	PostCompressed []byte
}

type RespondHookContext struct {
	*baseHook
	ResponseBody RespondBody
}

type RespondErrorBody struct {
	PreEncoded     error
	PostEncoded    []byte
	PostCompressed []byte
}

type RespondErrorHookContext struct {
	*baseHook
	ResponseBody RespondErrorBody
}

type RespondStreamBody struct {
	Value          ClonedBody
	IsCompressed   bool
	PostCompressed ClonedBody
}

type RespondStreamHookContext struct {
	*baseHook
	ResponseBody RespondStreamBody
}

type RespondHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodyStreamSize(respondContentType string, request *http.Request) int

	BeforeRespond(ctx *RespondContext, request *http.Request) *RespondContext
	RespondHook(ctx *RespondHookContext)
	RespondErrorHookContext(ctx *RespondErrorHookContext)
	RespondStreamHookContext(ctx *RespondStreamHookContext)
}

type (
	BeforeRespondFunc      = func(ctx *RespondContext, request *http.Request) *RespondContext
	ResponseHookFunc       = func(ctx *RespondHookContext)
	ResponseErrorHookFunc  = func(ctx *RespondErrorHookContext)
	ResponseStreamHookFunc = func(ctx *RespondStreamHookContext)
)

func (r *Responder) RegisterHook(hook RespondHook) {
	r.hooks = append(r.hooks, hook)
}
