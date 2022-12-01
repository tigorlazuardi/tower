package towerhttp

import (
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

// LoggingMiddleware is a middleware that enables logging the request and the response.
func LoggingMiddleware(logger ServerLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				body                   = r.Body
				requestBody ClonedBody = noopCloneBody{}
			)
			n := logger.ReceiveRequestBody(r)
			if n != 0 {
				clone := wrapClientBodyCloner(body, n, nil)
				body = clone
				requestBody = clone
			}
			capturer := newResponseCapture(w, r, logger)
			ctx := contextWithResponseCapture(r.Context(), capturer)
			r.Body = body
			r = r.WithContext(ctx)
			next.ServeHTTP(capturer, r)
			var caller tower.Caller
			if capturer.caller == nil {
				caller = tower.GetCaller(4)
			} else {
				caller = capturer.caller
			}
			t := capturer.tower
			if t == nil {
				t = tower.Global.Tower()
			}
			logger.Log(&ServerLoggerContext{
				Context:        ctx,
				Request:        r,
				RequestBody:    requestBody,
				ResponseStatus: capturer.status,
				ResponseHeader: capturer.w.Header(),
				ResponseBody:   capturer.body,
				Error:          capturer.writeError,
				Caller:         caller,
				Tower:          t,
			})
		})
	}
}
