package towerhttp

import (
	"bytes"
	"context"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
)

var interceptorKey struct{ key int } = struct{ key int }{777}

func contextWithLogger(ctx context.Context, interceptor *loggerInterceptor) context.Context {
	return context.WithValue(ctx, interceptorKey, interceptor)
}

func loggerFromContext(ctx context.Context) *loggerInterceptor {
	interceptor, _ := ctx.Value(interceptorKey).(*loggerInterceptor)
	return interceptor
}

type loggerInterceptor struct {
	request     *http.Request
	requestBody ClonedBody

	logger ServerLogger
}

type loggerContext struct {
	ctx            context.Context
	responseHeader http.Header
	responseStatus int
	responseBody   []byte
	err            error
	caller         tower.Caller
	tower          *tower.Tower
}

// receiveResponse will be called after the request is sent.
func (s *loggerInterceptor) log(ctx *loggerContext) {
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
		Tower:          ctx.tower,
	})
}

type loggerStreamContext struct {
	ctx         context.Context
	w           http.ResponseWriter
	contentType string
	body        io.Reader
	caller      tower.Caller
}

func (s *loggerInterceptor) logStream(ctx *loggerStreamContext) (wrappedBody io.Reader, wrappedRW responseCallback) {
	var clone ClonedBody = noopCloneBody{}
	wrappedBody = ctx.body
	size := s.logger.ReceiveResponseBodyStream(ctx.contentType, s.request)
	if size != 0 {
		b := wrapClientBodyCloner(wrappedBody, size, nil)
		wrappedBody = b
		clone = b
	}
	wrappedRW = newResponseCallback(ctx.ctx, ctx.w, func(statusCode int, size int, err error) {
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
