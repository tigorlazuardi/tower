package towerhttp

import (
	"io"
	"net/http"

	"github.com/tigorlazuardi/tower"
)

type RespondHookContext struct {
	Request        *http.Request
	RequestBody    ClonedBody
	ResponseStatus int
	ResponseHeader http.Header
	Body           struct {
		PreEncoded     any
		PostEncoded    []byte
		PostCompressed []byte
	}
	Caller tower.Caller
	Tower  *tower.Tower
	Error  error
}

type RespondErrorHookContext struct {
	Request        *http.Request
	RequestBody    ClonedBody
	ResponseStatus int
	ResponseHeader http.Header
	Body           struct {
		PreEncoded     error
		PostEncoded    []byte
		PostCompressed []byte
	}
	Caller tower.Caller
	Tower  *tower.Tower
	Error  error
}

type RespondStreamHookContext struct {
	Request        *http.Request
	RequestBody    ClonedBody
	ResponseStatus int
	ResponseHeader http.Header
	Body           struct {
		Value          ClonedBody
		IsCompressed   bool
		PostCompressed ClonedBody
	}
	Caller tower.Caller
	Tower  *tower.Tower
	Error  error
}

type hookRequest struct {
	requestBody  int
	responseBody int
}

type CloneRequest struct {
	hooks           map[RespondHook]*hookRequest
	requestReadMax  int
	responseReadMax int
}

func (cr *CloneRequest) ReadRequestBody(h RespondHook, size int) {
	hook, ok := cr.hooks[h]
	if !ok {
		newHook := &hookRequest{}
		hook = newHook
		cr.hooks[h] = hook
	}
	hook.requestBody = size
	if cr.requestReadMax > 0 && size > cr.requestReadMax {
		cr.requestReadMax = size
	}
}

type RespondHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodyStreamSize(contentType string, body io.Reader) int
	RespondHook(ctx *RespondHookContext)
	RespondErrorHookContext(ctx *RespondErrorHookContext)
	RespondStreamHookContext(ctx *RespondStreamHookContext)
}
