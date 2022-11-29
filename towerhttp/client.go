package towerhttp

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client interface {
	Doer
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
	CloseIdleConnections()
}

type getRequester interface {
	Get(url string) (*http.Response, error)
}

type headRequester interface {
	Head(url string) (*http.Response, error)
}

type postRequester interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

type postFormRequester interface {
	PostForm(string, url.Values) (*http.Response, error)
}

type closeIdleConnections interface {
	CloseIdleConnections()
}

type HTTPClient struct {
	inner  Doer
	logger ClientLogger
}

func (H HTTPClient) Do(request *http.Request) (*http.Response, error) {
	if request.Body != nil {
		bodyRequest := H.logger.ReceiveRequestBody(request)
		if bodyRequest != 0 {
			request.Body = wrapClientBodyCloner(request.Body, bodyRequest)
		}
	}
	// TODO implement better API
	return H.inner.Do(request)
}

func (H HTTPClient) Get(url string) (*http.Response, error) {
	if get, ok := H.inner.(getRequester); ok {
		return get.Get(url)
	}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return H.Do(req)
}

func (H HTTPClient) Head(url string) (*http.Response, error) {
	if head, ok := H.inner.(headRequester); ok {
		return head.Head(url)
	}
	req, err := http.NewRequest(http.MethodHead, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return H.Do(req)
}

func (H HTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	if post, ok := H.inner.(postRequester); ok {
		return post.Post(url, contentType, body)
	}
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return H.Do(req)
}

func (H HTTPClient) PostForm(s string, values url.Values) (*http.Response, error) {
	if postForm, ok := H.inner.(postFormRequester); ok {
		return postForm.PostForm(s, values)
	}
	return H.Post(s, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
}

func (H HTTPClient) CloseIdleConnections() {
	if closer, ok := H.inner.(closeIdleConnections); ok {
		closer.CloseIdleConnections()
	}
}
