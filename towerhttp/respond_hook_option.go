package towerhttp

import (
	"net/http"
)

type RespondHookOptionGroup struct{}

type RespondHookOption interface {
	apply(*respondHook)
}

type RespondHookOptionFunc func(*respondHook)

func (r RespondHookOptionFunc) apply(hook *respondHook) {
	r(hook)
}

type FilterRequest = func(*http.Request) bool

type FilterRespond = func(respondContentType string, r *http.Request) bool

var _ RespondHook = (*respondHook)(nil)

type respondHook struct {
	readRequestLimit    int
	readRespondLimit    int
	filterRequest       FilterRequest
	filterRespondStream FilterRespond
	beforeRespond       BeforeRespondFunc
	onRespond           ResponseHookFunc
	onRespondError      ResponseErrorHookFunc
	onRespondStream     ResponseStreamHookFunc
}

func NewRespondHook(opts ...RespondHookOption) RespondHook {
	r := &respondHook{}
	for _, opt := range opts {
		opt.apply(r)
	}
	return r
}

func (r2 respondHook) AcceptRequestBodySize(r *http.Request) int {
	if r2.filterRequest != nil && r2.filterRequest(r) {
		return r2.readRequestLimit
	}
	return 0
}

func (r2 respondHook) AcceptResponseBodyStreamSize(contentType string, request *http.Request) int {
	if r2.filterRequest != nil && r2.filterRespondStream(contentType, request) {
		return r2.readRespondLimit
	}
	return 0
}

func (r2 respondHook) BeforeRespond(ctx *RespondContext, request *http.Request) *RespondContext {
	if r2.beforeRespond == nil {
		return ctx
	}
	return r2.beforeRespond(ctx, request)
}

func (r2 respondHook) RespondHook(ctx *RespondHookContext) {
	if r2.onRespond != nil {
		r2.onRespond(ctx)
	}
}

func (r2 respondHook) RespondErrorHookContext(ctx *RespondErrorHookContext) {
	if r2.onRespondError != nil {
		r2.onRespondError(ctx)
	}
}

func (r2 respondHook) RespondStreamHookContext(ctx *RespondStreamHookContext) {
	if r2.onRespondStream != nil {
		r2.onRespondStream(ctx)
	}
}

func (RespondHookOptionGroup) ReadRequestBodyLimit(limit int) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.readRequestLimit = limit
	})
}

func (RespondHookOptionGroup) ReadRespondBodyStreamLimit(limit int) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.readRespondLimit = limit
	})
}

func (RespondHookOptionGroup) FilterRequest(filter FilterRequest) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.filterRequest = filter
	})
}

func (RespondHookOptionGroup) FilterRespondStream(filter FilterRespond) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.filterRespondStream = filter
	})
}

func (RespondHookOptionGroup) BeforeRespond(before BeforeRespondFunc) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.beforeRespond = before
	})
}

func (RespondHookOptionGroup) OnRespond(on ResponseHookFunc) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.onRespond = on
	})
}

func (RespondHookOptionGroup) OnRespondError(on ResponseErrorHookFunc) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.onRespondError = on
	})
}

func (RespondHookOptionGroup) OnRespondStream(on ResponseStreamHookFunc) RespondHookOption {
	return RespondHookOptionFunc(func(r *respondHook) {
		r.onRespondStream = on
	})
}
