package towerhttp

import (
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
			ctx := contextWithLogger(r.Context(), &loggerInterceptor{
				request:     r,
				requestBody: requestBody,
				logger:      logger,
			})
			r.Body = body
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
