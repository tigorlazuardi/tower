package towerhttp

import (
	"io"
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

type RoundTrip struct {
	inner http.RoundTripper
	hook  RoundTripHook
	tower *tower.Tower
}

type bodyCloseHook struct {
	io.ReadCloser
	cb func()
}

func (bch *bodyCloseHook) Close() error {
	err := bch.ReadCloser.Close()
	bch.cb()
	return err
}

func (rt *RoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
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

// NewRoundTrip creates a new instance of http.RoundTripper implementation that wraps around the default http.RoundTripper.
// It provides logging support using tower engine.
func NewRoundTrip(opts ...RoundTripOption) *RoundTrip {
	roundtrip := &RoundTrip{inner: http.DefaultTransport, hook: NewRoundTripHook(), tower: tower.Global.Tower()}
	for _, v := range opts {
		v.apply(roundtrip)
	}
	return roundtrip
}

// WrapRoundTripper wraps the given http.RoundTripper with towerhttp.RoundTrip implementation to support logging with
// tower engine.
func WrapRoundTripper(rt http.RoundTripper, opts ...RoundTripOption) *RoundTrip {
	roundtrip := &RoundTrip{inner: rt, hook: NewRoundTripHook(), tower: tower.Global.Tower()}
	for _, v := range opts {
		v.apply(roundtrip)
	}
	return roundtrip
}

// WrapHTTPClient wraps the given http.Client with towerhttp.RoundTrip implementation to support logging with
// tower engine.
func WrapHTTPClient(client *http.Client, opts ...RoundTripOption) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client.Transport = WrapRoundTripper(transport, opts...)
	return client
}

type RoundTripOption interface {
	apply(*RoundTrip)
}

type (
	RoundTripOptionFunc    func(*RoundTrip)
	RoundTripOptionBuilder []RoundTripOption
)

func (rto RoundTripOptionFunc) apply(rt *RoundTrip) {
	rto(rt)
}

func (rtob RoundTripOptionBuilder) apply(rt *RoundTrip) {
	for _, v := range rtob {
		v.apply(rt)
	}
}
