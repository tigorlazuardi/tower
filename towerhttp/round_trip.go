package towerhttp

import (
	"net/http"

	"github.com/tigorlazuardi/tower"
)

type RoundTripContext struct {
	Request     *http.Request
	RequestBody ClonedBody
	// Response is nil if there is an error to get HTTP Response.
	Response     *http.Response
	ResponseBody ClonedBody
	Error        error
	Tower        *tower.Tower
}

type RoundTripHook interface {
	AcceptRequestBodySize(r *http.Request) int
	AcceptResponseBodySize(req *http.Request, res *http.Response) int
	ExecuteHook(ctx *RoundTripContext)
}

type (
	RoundTripFilterRequest  = func(*http.Request) bool
	RoundTripFilterResponse = func(*http.Request, *http.Response) bool
	RoundTripExecuteHook    = func(*RoundTripContext)
)

type roundTripHook struct {
	readRespondLimit int
	readRequestLimit int
	filterRequest    RoundTripFilterRequest
	filterResponse   RoundTripFilterResponse
	log              RoundTripExecuteHook
}

type RoundTripper struct {
	inner http.RoundTripper
	hook  RoundTripHook
	tower *tower.Tower
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		reqBody ClonedBody = NoopCloneBody{}
		resBody ClonedBody = NoopCloneBody{}
	)
	req = req.Clone(req.Context())
	wantReqBody := rt.hook.AcceptRequestBodySize(req)
	if wantReqBody != 0 {
		reqBodyClone := wrapBodyCloner(req.Body, wantReqBody)
		req.Body = reqBodyClone
		reqBody = reqBodyClone
	}
	res, err := rt.inner.RoundTrip(req)
	ctx := &RoundTripContext{
		Request:      req,
		RequestBody:  reqBody,
		Response:     res,
		ResponseBody: resBody,
		Error:        err,
		Tower:        rt.tower,
	}
	if res != nil {
		wantResBody := rt.hook.AcceptResponseBodySize(req, res)
		if wantResBody != 0 {
			resBodyClone := wrapBodyCloner(res.Body, wantResBody)
			resBody = resBodyClone
			res.Body = &bodyCloseHook{ReadCloser: resBodyClone, cb: func() {
				rt.hook.ExecuteHook(ctx)
			}}
		}
	} else {
		rt.hook.ExecuteHook(ctx)
	}

	return res, err
}

func (rth *roundTripHook) AcceptRequestBodySize(r *http.Request) int {
	if rth.filterRequest(r) {
		return rth.readRequestLimit
	}
	return 0
}

func (rth *roundTripHook) AcceptResponseBodySize(req *http.Request, res *http.Response) int {
	if rth.filterResponse(req, res) {
		return rth.readRespondLimit
	}
	return 0
}

func (rth *roundTripHook) ExecuteHook(ctx *RoundTripContext) {
	rth.log(ctx)
}

func WrapRoundTripper(rt http.RoundTripper) *RoundTripper {
	// TODO: make this configurable
	return &RoundTripper{inner: rt}
}

func WrapHTTPClient(client *http.Client) *http.Client {
	// TODO: make this configurable
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = WrapRoundTripper(client.Transport)
	return client
}
