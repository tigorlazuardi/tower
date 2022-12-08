package towerhttp

import (
	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponder_RespondStream(t *testing.T) {
	towerGen := func() (*tower.Tower, *tower.TestingJSONLogger) {
		logger := tower.NewTestingJSONLogger()
		tow := tower.NewTower(tower.Service{
			Name:        "test",
			Environment: "test",
			Type:        "test",
		})
		tow.SetLogger(logger)
		return tow, logger
	}
	type fields struct {
		encoder          Encoder
		transformer      BodyTransformer
		errorTransformer ErrorBodyTransformer
		compressor       Compressor
		callerDepth      int
	}
	type respond struct {
		contentType string
		body        io.Reader
		opts        []RespondOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    respond
		request func(server *httptest.Server) *http.Request
		test    func(t *testing.T, logger *tower.TestingJSONLogger, resp *http.Response)
	}{
		{
			name: "common pattern",
			fields: fields{
				encoder:          nil,
				transformer:      nil,
				errorTransformer: SimpleErrorTransformer{},
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			args: respond{
				contentType: "text/plain; charset=utf-8",
				body:        strings.NewReader("hello world"),
				opts:        nil,
			},
			request: func(server *httptest.Server) *http.Request {
				req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
				return req
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger, resp *http.Response) {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/plain; charset=utf-8" {
					t.Errorf("expected content type to be %s, got %s", "text/plain; charset=utf-8", contentType)
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != "hello world" {
					t.Errorf("expected body to be %s, got %s", "hello world", string(body))
				}
				logs := logger.String()
				if !strings.Contains(logs, "respond_stream_test.go") {
					t.Errorf("expected log caller to contain %s, got %s", "respond_stream_test.go", logs)
				}
				j := jsonassert.New(t)
				wantLog := `
				{
					"time": "<<PRESENCE>>",
					"code": 200,
					"message": "GET / HTTP/1.1",
					"caller": "<<PRESENCE>>",
					"level": "info",
					"service": {
						"name": "test",
						"environment": "test",
						"type": "test"
					},
					"context": {
						"request": {
							"headers": {
								"Accept-Encoding": [
									"gzip"
								],
								"User-Agent": [
									"Go-http-client/1.1"
								]
							},
							"method": "GET",
							"url": "%s/"
						},
						"response": {
							"body": "hello world",
							"headers": {
								"Content-Type": [
									"text/plain; charset=utf-8"
								]
							},
							"status": 200
						}
					}
				}`
				j.Assertf(logs, wantLog, resp.Request.Host)
			},
		},
		{
			name: "handled nil body",
			fields: fields{
				encoder:          nil,
				transformer:      nil,
				errorTransformer: SimpleErrorTransformer{},
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			args: respond{
				contentType: "",
				body:        nil,
				opts:        nil,
			},
			request: func(server *httptest.Server) *http.Request {
				req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
				return req
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger, resp *http.Response) {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "" {
					t.Errorf("expected content type to be empty")
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != "" {
					t.Errorf("expected body to be empty")
				}
				logs := logger.String()
				if !strings.Contains(logs, "respond_stream_test.go") {
					t.Errorf("expected log caller to contain %s, got %s", "respond_stream_test.go", logs)
				}
				j := jsonassert.New(t)
				wantLog := `
				{
					"time": "<<PRESENCE>>",
					"code": 200,
					"message": "GET / HTTP/1.1",
					"caller": "<<PRESENCE>>",
					"level": "info",
					"service": {
						"name": "test",
						"environment": "test",
						"type": "test"
					},
					"context": {
						"request": {
							"headers": {
								"Accept-Encoding": [
									"gzip"
								],
								"User-Agent": [
									"Go-http-client/1.1"
								]
							},
							"method": "GET",
							"url": "%s/"
						},
						"response": {
							"status": 200
						}
					}
				}`
				j.Assertf(logs, wantLog, resp.Request.Host)
			},
		},
		{
			name: "handled http.NoBody",
			fields: fields{
				encoder:          nil,
				transformer:      nil,
				errorTransformer: SimpleErrorTransformer{},
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			args: respond{
				contentType: "",
				body:        http.NoBody,
				opts: []RespondOption{
					Option.Respond().StatusCode(http.StatusNoContent),
					Option.Respond().AddCallerSkip(0),
					Option.Respond().Transformer(NoopBodyTransform{}),
					Option.Respond().Encoder(NewJSONEncoder()),
					Option.Respond().Compressor(NoCompression{}),
					Option.Respond().CallerSkip(2),
					Option.Respond().ErrorTransformer(SimpleErrorTransformer{}),
				},
			},
			request: func(server *httptest.Server) *http.Request {
				req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
				return req
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger, resp *http.Response) {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "" {
					t.Errorf("expected content type to be empty")
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != "" {
					t.Errorf("expected body to be empty")
				}
				logs := logger.String()
				if !strings.Contains(logs, "respond_stream_test.go") {
					t.Errorf("expected log caller to contain %s, got %s", "respond_stream_test.go", logs)
				}
				j := jsonassert.New(t)
				wantLog := `
				{
					"time": "<<PRESENCE>>",
					"code": 204,
					"message": "GET / HTTP/1.1",
					"caller": "<<PRESENCE>>",
					"level": "info",
					"service": {
						"name": "test",
						"environment": "test",
						"type": "test"
					},
					"context": {
						"request": {
							"headers": {
								"Accept-Encoding": [
									"gzip"
								],
								"User-Agent": [
									"Go-http-client/1.1"
								]
							},
							"method": "GET",
							"url": "%s/"
						},
						"response": {
							"status": 204
						}
					}
				}`
				j.Assertf(logs, wantLog, resp.Request.Host)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tow, logger := towerGen()
			defer func() {
				if t.Failed() {
					logger.PrettyPrint()
				}
			}()
			r := Responder{
				encoder:          tt.fields.encoder,
				transformer:      tt.fields.transformer,
				errorTransformer: tt.fields.errorTransformer,
				tower:            tow,
				compressor:       tt.fields.compressor,
				callerDepth:      tt.fields.callerDepth,
				hooks:            []RespondHook{NewLoggerHook()},
			}
			handler := func(writer http.ResponseWriter, request *http.Request) {
				r.RespondStream(writer, request, tt.args.contentType, tt.args.body, tt.args.opts...)
			}
			server := httptest.NewServer(r.RequestBodyCloner()(http.HandlerFunc(handler)))
			defer server.Close()
			req := tt.request(server)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			tt.test(t, logger, resp)
		})
	}
}
