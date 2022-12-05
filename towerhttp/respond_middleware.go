package towerhttp

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func (r Responder) RequestBodyCloner() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			var count int
			for _, hook := range r.hooks {
				accept := hook.AcceptRequestBodySize(request)
				// A hook requests all body to be read. We will read all body and stop looking at other hooks.
				if accept < 0 {
					count = accept
					break
				}
				if count >= 0 && accept > count {
					count = accept
				}
			}
			if count != 0 {
				cloner := wrapBodyCloner(request.Body, count)
				request.Body = cloner
				ctx := contextWithClonedBody(request.Context(), cloner)
				request = request.WithContext(ctx)
			}
			next.ServeHTTP(writer, request)
		})
	}
}
