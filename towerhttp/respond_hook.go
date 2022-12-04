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

type CheckedHook struct {
	hooks map[RespondHook]*hookRequest
}

type RespondHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodyStreamSize(contentType string, body io.Reader) int
	RespondHook(ctx *RespondHookContext)
	RespondErrorHookContext(ctx *RespondErrorHookContext)
	RespondStreamHookContext(ctx *RespondStreamHookContext)
}
