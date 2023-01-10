package towerhttp

import (
	"context"
	"net/http"
)

func (t TowerClient) Get(url string) (*http.Response, error) {
	if ec, ok := t.inner.(ExtendedClient); ok {
		return t.getExtended(ec, url)
	}
	return t.get(url)
}

func (t TowerClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.inner.Do(req)
	if resp != nil {
		wantRespBody := t.hook.AcceptResponseBodySize(req, resp)
		rc := wrapBodyCloner(resp.Body, wantRespBody)
		resp.Body = &bodyCloseHook{ReadCloser: rc, cb: func() {
			ctx := req.Context()
			t.hook.ExecuteHook(&ClientHookContext{
				Context:      ctx,
				Request:      req,
				RequestBody:  NoopCloneBody{},
				Response:     resp,
				ResponseBody: rc,
				Error:        err,
			})
		}}
		return resp, err
	}
	t.hook.ExecuteHook(&ClientHookContext{
		Context:      req.Context(),
		Request:      req,
		RequestBody:  NoopCloneBody{},
		Response:     resp,
		ResponseBody: NoopCloneBody{},
		Error:        err,
	})
	return resp, err
}

func (t TowerClient) getExtended(client ExtendedClient, url string) (*http.Response, error) {
	ctx := context.Background()
	resp, err := client.Get(url)
	var req *http.Request
	if resp != nil {
		req = resp.Request
		wantRespBody := t.hook.AcceptResponseBodySize(req, resp)
		rc := wrapBodyCloner(resp.Body, wantRespBody)
		resp.Body = &bodyCloseHook{ReadCloser: rc, cb: func() {
			if req != nil {
				ctx = req.Context()
			}
			t.hook.ExecuteHook(&ClientHookContext{
				Context:      ctx,
				Request:      req,
				RequestBody:  NoopCloneBody{},
				Response:     resp,
				ResponseBody: rc,
				Error:        err,
			})
		}}
		return resp, err
	}
	t.hook.ExecuteHook(&ClientHookContext{
		Context:      ctx,
		Request:      req,
		RequestBody:  NoopCloneBody{},
		Response:     resp,
		Error:        err,
		ResponseBody: NoopCloneBody{},
	})
	return resp, err
}
