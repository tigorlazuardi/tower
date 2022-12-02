package towerhttp

import (
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type (
	FilterServerRequestFunc  = func(r *http.Request) bool
	FilterServerResponseFunc = func(contentType string, req *http.Request) bool
)


//type ServerLoggerOptionGroup interface {
//	// AcceptRequestBody sets whether the request body should be cloned.
//	AcceptRequestBody(requestFunc FilterServerRequestFunc) ServerLoggerOption
//	// AcceptResponseBodyStream sets whether the response body should be cloned.
//	//
//	// Non-streaming response body will always be received by the logger.
//	AcceptResponseBodyStream(responseFunc FilterServerResponseFunc) ServerLoggerOption
//	// Notify enables tower Messengers to be notified when the request-response trip is done.
//	Notify(enable bool, options ...tower.MessageOption) ServerLoggerOption
//}

type ServerLoggerOption interface {
	apply(opts *serverLoggerOpts)
}

type ServerLoggerOptionFunc func(opts *serverLoggerOpts)

func (s ServerLoggerOptionFunc) apply(opts *serverLoggerOpts) {
	s(opts)
}

type serverLoggerOpts struct {
	limit          int
	requestFilter  FilterServerRequestFunc
	responseFilter FilterServerResponseFunc
	notify         bool
	notifyOption   []tower.MessageOption
}

type ServerLoggerOptionGroup struct{}

// Limit sets the maximum number of bytes to be read from the request or response body.
// 0 means no body will be cloned.
// -1 means the entire body will be cloned.
func (ServerLoggerOptionGroup) Limit(i int) ServerLoggerOption {
	return ServerLoggerOptionFunc(func(opts *serverLoggerOpts) {
		opts.limit = i
	})
}

func (ServerLoggerOptionGroup) AcceptRequestBody(f FilterServerRequestFunc) ServerLoggerOption {
	return ServerLoggerOptionFunc(func(opts *serverLoggerOpts) {
		opts.requestFilter = f
	})
}

// AcceptResponseBodyStream sets whether the response body should be cloned.
//
// Non-streaming response body will always be received by the logger.
func (ServerLoggerOptionGroup) AcceptResponseBodyStream(f FilterServerResponseFunc) ServerLoggerOption {
	return ServerLoggerOptionFunc(func(opts *serverLoggerOpts) {
		opts.responseFilter = f
	})
}

// Notify enables tower Messengers to be notified when the request-response trip is done.
func (ServerLoggerOptionGroup) Notify(enable bool, options ...tower.MessageOption) ServerLoggerOption {
	return ServerLoggerOptionFunc(func(opts *serverLoggerOpts) {
		opts.notify = enable
		opts.notifyOption = options
	})
}