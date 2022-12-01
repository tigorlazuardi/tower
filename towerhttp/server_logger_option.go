package towerhttp

import (
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type (
	FilterRequestFunc  = func(r *http.Request) bool
	FilterResponseFunc = func(contentType string, req *http.Request) bool
)

type ServerLoggerOptionGroup interface {
	// Limit sets the maximum number of bytes to be read from the request or response body.
	// 0 means no body will be cloned.
	// -1 means the entire body will be cloned.
	Limit(limit int) ServerLoggerOption
	// AcceptRequestBody sets whether the request body should be cloned.
	AcceptRequestBody(requestFunc FilterRequestFunc) ServerLoggerOption
	// AcceptResponseBodyStream sets whether the response body should be cloned.
	//
	// Non-streaming response body will always be received by the logger.
	AcceptResponseBodyStream(responseFunc FilterResponseFunc) ServerLoggerOption
	// Notify enables tower Messengers to be notified when the request-response trip is done.
	Notify(enable bool, options ...tower.MessageOption) ServerLoggerOption
}

type ServerLoggerOption interface {
	apply(opts *serverLoggerOpts)
}

type ServerLoggerOptionFunc func(opts *serverLoggerOpts)

func (s ServerLoggerOptionFunc) apply(opts *serverLoggerOpts) {
	s(opts)
}

type serverLoggerOpts struct {
	limit          int
	requestFilter  FilterRequestFunc
	responseFilter FilterResponseFunc
	notify         bool
	notifyOption   []tower.MessageOption
}
