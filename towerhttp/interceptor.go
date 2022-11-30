package towerhttp

import (
	"bytes"
	"context"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
)

type serverInterceptor struct {
	request     *http.Request
	requestBody ClonedBody

	logger ServerLogger
}

type serverInterceptorContext struct {
	ctx            context.Context
	responseHeader http.Header
	responseStatus int
	responseBody   []byte
	err            error
	caller         tower.Caller
}

// receiveResponse will be called after the request is sent.
func (s *serverInterceptor) log(ctx *serverInterceptorContext) {
	clone := bytes.NewBuffer(ctx.responseBody)
	sb := io.NopCloser(clone)

	respBody := clientBodyCloner{
		ReadCloser: sb,
		clone:      clone,
		limit:      -1,
		callback:   nil,
	}

	s.logger.Log(&ServerLoggerContext{
		Context:        ctx.ctx,
		Request:        s.request,
		RequestBody:    s.requestBody,
		ResponseStatus: ctx.responseStatus,
		ResponseHeader: ctx.responseHeader,
		ResponseBody:   respBody,
		Error:          ctx.err,
		Caller:         ctx.caller,
	})
}

type serverInterceptorStreamContext struct {
	ctx         context.Context
	w           http.ResponseWriter
	contentType string
	body        io.Reader
	caller      tower.Caller
}

func (s *serverInterceptor) logStream(ctx *serverInterceptorStreamContext) (wrappedBody io.Reader, wrappedRW responseCallback) {
	var clone ClonedBody = noopCloneBody{}
	wrappedBody = ctx.body
	size := s.logger.ReceiveResponseBodyStream(ctx.contentType, s.request)
	if size != 0 {
		b := wrapClientBodyCloner(wrappedBody, size, nil)
		wrappedBody = b
		clone = b
	}
	wrappedRW = newResponseCallback(ctx.w, func(statusCode int, size int, err error) {
		s.logger.Log(&ServerLoggerContext{
			Context:        ctx.ctx,
			Request:        s.request,
			RequestBody:    s.requestBody,
			ResponseStatus: statusCode,
			ResponseHeader: ctx.w.Header(),
			ResponseBody:   clone,
			Error:          err,
			Caller:         ctx.caller,
		})
	})
	return
}
