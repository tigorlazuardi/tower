package towerhttp

// type Doer interface {
// 	Do(*http.Request) (*http.Response, error)
// }
//
// type Client interface {
// 	Doer
// 	Get(url string) (*http.Response, error)
// 	Head(url string) (*http.Response, error)
// 	Post(url string, contentType string, body io.Reader) (*http.Response, error)
// 	PostForm(string, url.Values) (*http.Response, error)
// 	CloseIdleConnections()
// }
//
// type getRequester interface {
// 	Get(url string) (*http.Response, error)
// }
//
// type headRequester interface {
// 	Head(url string) (*http.Response, error)
// }
//
// type postRequester interface {
// 	Post(url string, contentType string, body io.Reader) (*http.Response, error)
// }
//
// type postFormRequester interface {
// 	PostForm(string, url.Values) (*http.Response, error)
// }
//
// type closeIdleConnections interface {
// 	CloseIdleConnections()
// }
//
// type HTTPClient struct {
// 	inner       Doer
// 	logger      ClientLogger
// 	callerDepth int
// }
//
// func (H *HTTPClient) SetCallerDepth(callerDepth int) {
// 	H.callerDepth = callerDepth
// }
//
// func (H HTTPClient) Do(request *http.Request) (*http.Response, error) {
// 	caller := tower.GetCaller(H.callerDepth)
// 	var reqBody ClonedBody = NoopCloneBody{}
// 	if request.Body != nil {
// 		bodyRequest := H.logger.ReceiveRequestBody(request)
// 		if bodyRequest != 0 {
// 			clone := wrapBodyCloner(request.Body, bodyRequest)
// 			reqBody = clone
// 			request.Body = clone
// 		}
// 	}
// 	ctx := &ClientRequestContext{
// 		Context:      request.Context(),
// 		Request:      request,
// 		RequestBody:  reqBody,
// 		ResponseBody: NoopCloneBody{},
// 		Caller:       caller,
// 	}
// 	resp, err := H.inner.Do(request)
// 	ctx.Response = resp
// 	ctx.Error = err
// 	if err != nil {
// 		H.logger.Log(ctx)
// 		return resp, err
// 	}
// 	bodyResponse := H.logger.ReceiveResponseBody(request, resp)
// 	if bodyResponse != 0 {
// 		clone := wrapBodyCloner(resp.Body, bodyResponse)
// 		clone.onClose(func(err error) {
// 			H.logger.Log(ctx)
// 		})
// 		resp.Body = clone
// 		ctx.ResponseBody = clone
// 	}
// 	return resp, err
// }
//
// func (H HTTPClient) Get(url string) (*http.Response, error) {
// 	if get, ok := H.inner.(getRequester); ok {
// 		return get.Get(url)
// 	}
// 	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return H.Do(req)
// }
//
// func (H HTTPClient) Head(url string) (*http.Response, error) {
// 	if head, ok := H.inner.(headRequester); ok {
// 		return head.Head(url)
// 	}
// 	req, err := http.NewRequest(http.MethodHead, url, http.NoBody)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return H.Do(req)
// }
//
// func (H HTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
// 	if post, ok := H.inner.(postRequester); ok {
// 		return post.Post(url, contentType, body)
// 	}
// 	req, err := http.NewRequest(http.MethodPost, url, body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("Content-Type", contentType)
// 	return H.Do(req)
// }
//
// func (H HTTPClient) PostForm(s string, values url.Values) (*http.Response, error) {
// 	if postForm, ok := H.inner.(postFormRequester); ok {
// 		return postForm.PostForm(s, values)
// 	}
// 	return H.Post(s, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
// }
//
// func (H HTTPClient) CloseIdleConnections() {
// 	if closer, ok := H.inner.(closeIdleConnections); ok {
// 		closer.CloseIdleConnections()
// 	}
// }
//
// // WrapClient wraps http client that implements towerhttp.Doer.
// func WrapClient(client Doer) Client {
// 	return &HTTPClient{inner: client, logger: NewTowerClientLogger(tower.Global.Tower()), callerDepth: 2}
// }
