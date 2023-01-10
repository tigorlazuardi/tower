package towerhttp

import (
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"net/url"
)

type bodyCloseHook struct {
	io.ReadCloser
	cb func()
}

func (hook *bodyCloseHook) Close() error {
	defer hook.cb()
	return hook.ReadCloser.Close()
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type ExtendedClient interface {
	Client
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
}

type TowerClient struct {
	inner Client
	tower *tower.Tower
	hook  ClientHook
}

func (t TowerClient) Do(request *http.Request) (*http.Response, error) {
	return t.inner.Do(request)
}

func (t TowerClient) Head(url string) (*http.Response, error) {
	if ec, ok := t.inner.(ExtendedClient); ok {
		return ec.Head(url)
	}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return t.inner.Do(req)
}

func (t TowerClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	if ec, ok := t.inner.(ExtendedClient); ok {
		return ec.Post(url, contentType, body)
	}
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return t.inner.Do(req)
}

func (t TowerClient) PostForm(s string, values url.Values) (*http.Response, error) {
	if ec, ok := t.inner.(ExtendedClient); ok {
		return ec.PostForm(s, values)
	}
	req, err := http.NewRequest(http.MethodPost, s, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return t.inner.Do(req)
}
