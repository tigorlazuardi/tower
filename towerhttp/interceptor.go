package towerhttp

import "net/http"

var interceptorKey = struct{ key string }{"towerhttp.interceptor"}

type interceptor struct {
	request      *http.Request
	requestBody  ClonedBody
	response     *http.ResponseWriter
	responseBody ClonedBody
}
