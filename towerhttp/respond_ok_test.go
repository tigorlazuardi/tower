package towerhttp

import (
	"bytes"
	"context"
	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type statusCreatedBody struct{}

func (s statusCreatedBody) MarshalJSON() ([]byte, error) {
	return []byte(`{"status":"created"}`), nil
}

func (s statusCreatedBody) HTTPCode() int {
	return http.StatusCreated
}

func TestResponder_Respond(t *testing.T) {
	type fields struct {
		encoder          Encoder
		transformer      BodyTransformer
		errorTransformer ErrorBodyTransformer
		compressor       Compressor
		callerDepth      int
	}
	type gen struct {
		server func(*Responder, Middleware) *httptest.Server
		tower  func(logger tower.Logger) *tower.Tower
	}
	towerGen := func(logger tower.Logger) *tower.Tower {
		t := tower.NewTower(tower.Service{
			Name:        "responder-test",
			Environment: "testing",
			Type:        "unit-test",
		})
		t.SetLogger(logger)
		return t
	}
	tests := []struct {
		name   string
		fields fields
		gen    gen
		test   func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger)
	}{
		{
			name: "normal state",
			fields: fields{
				encoder:          NewJSONEncoder(),
				transformer:      NoopBodyTransform{},
				errorTransformer: nil,
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						body := map[string]string{"ok": "ok"}
						responder.Respond(request.Context(), writer, body)
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected content type %s, got %s", "application/json", resp.Header.Get("Content-Type"))
				}
				if resp.Header.Get("Content-Encoding") != "" {
					t.Errorf("Expected content encoding to be empty, got %s", resp.Header.Get("Content-Encoding"))
				}
				want := `{"ok":"ok"}`
				lenWant := strconv.Itoa(len(want) + 1)
				if resp.Header.Get("Content-Length") != lenWant {
					t.Errorf("Expected content length '%d', got '%s'", len(want)+1, resp.Header.Get("Content-Length"))
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Failed to read body: %s", err.Error())
					return
				}
				body = bytes.TrimSpace(body)
				if string(body) != want {
					t.Errorf("Expected body %s len(%d), got %s len(%d)", want, len(want), string(body), len(body))
				}
				if len(logger.Bytes()) == 0 {
					t.Errorf("Expected logger to be called, got empty log")
				}
				j := jsonassert.New(t)
				jsonStr := logger.String()
				j.Assertf(jsonStr, `
				{
					"time": "<<PRESENCE>>",
					"level": "info",
					"message": "GET /",
					"service": {
						"name": "responder-test",
						"environment": "testing",
						"type": "unit-test"
					},
					"caller": "<<PRESENCE>>",
					"context": {
						"request": {
							"method": "GET",
							"url": "%s/",
							"headers": {
								"Accept-Encoding": ["gzip"],
								"User-Agent": ["Go-http-client/1.1"]
							}
						},
						"response": {
							"status": 200,
							"headers": {
								"Content-Length": ["%s"],
								"Content-Type": ["application/json"]
							},
							"body": %s
						}
					}
				}`, resp.Request.Host, lenWant, want)
				substr := "tower/towerhttp/respond_ok_test.go"
				if !strings.Contains(jsonStr, "tower/towerhttp/respond_ok_test.go") {
					t.Errorf("Expected caller to be present to contains '%s'", substr)
				}
			},
		},
		{
			name: "http.NoBody",
			fields: fields{
				encoder:          NewJSONEncoder(),
				transformer:      NoopBodyTransform{},
				errorTransformer: nil,
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						responder.Respond(request.Context(), writer, http.NoBody, Option.Respond().StatusCode(http.StatusNoContent))
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusNoContent {
					t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "" {
					t.Errorf("Expected content type to be empty, but got %s", resp.Header.Get("Content-Type"))
				}
				if resp.Header.Get("Content-Encoding") != "" {
					t.Errorf("Expected content encoding to be empty, got %s", resp.Header.Get("Content-Encoding"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) != 0 {
					t.Errorf("Expected body to be empty, got %s", string(body))
					return
				}
				log := `
				{
				  "time": "<<PRESENCE>>",
				  "message": "GET /",
				  "caller": "<<PRESENCE>>",
				  "level": "info",
				  "service": {
					"name": "responder-test",
					"environment": "testing",
					"type": "unit-test"
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
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host)
			},
		},
		{
			name: "nil body - on default BodyTransform",
			fields: fields{
				encoder:          NewJSONEncoder(),
				transformer:      NoopBodyTransform{},
				errorTransformer: nil,
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						responder.Respond(request.Context(), writer, nil)
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "" {
					t.Errorf("Expected content type to be empty, but got %s", resp.Header.Get("Content-Type"))
				}
				if resp.Header.Get("Content-Encoding") != "" {
					t.Errorf("Expected content encoding to be empty, got %s", resp.Header.Get("Content-Encoding"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) != 0 {
					t.Errorf("Expected body to be empty, got %s", string(body))
					return
				}
				log := `
				{
				  "time": "<<PRESENCE>>",
				  "message": "GET /",
				  "caller": "<<PRESENCE>>",
				  "level": "info",
				  "service": {
					"name": "responder-test",
					"environment": "testing",
					"type": "unit-test"
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
					  "status": %d
					}
				  }
				}`
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host, resp.StatusCode)
			},
		},
		{
			name: "nil body - custom BodyTransform",
			fields: fields{
				encoder: NewJSONEncoder(),
				transformer: BodyTransformFunc(func(ctx context.Context, input any) any {
					return map[string]any{
						"message": "custom body transform",
						"data":    input,
					}
				}),
				errorTransformer: nil,
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						responder.Respond(request.Context(), writer, nil)
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected content type to be 'application/json', but got '%s'", resp.Header.Get("Content-Type"))
				}
				if resp.Header.Get("Content-Encoding") != "" {
					t.Errorf("Expected content encoding to be empty, got %s", resp.Header.Get("Content-Encoding"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) == 0 {
					t.Errorf("Expected body to be not empty")
					return
				}
				log := `
				{
				  "time": "<<PRESENCE>>",
				  "message": "GET /",
				  "caller": "<<PRESENCE>>",
				  "level": "info",
				  "service": {
					"name": "responder-test",
					"environment": "testing",
					"type": "unit-test"
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
					  "status": %d,
                      "body": {
						"message": "custom body transform",
						"data": null
					  },
                      "headers": {
						"Content-Type": [
						  "application/json"
						],
						"Content-Length": [
						  "%s"
						]
					  }
					}
				  }
				}`
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host, resp.StatusCode, resp.Header.Get("Content-Length"))
			},
		},
		{
			name: "nil body - custom BodyTransform - From option override",
			fields: fields{
				encoder: NewJSONEncoder(),
				transformer: BodyTransformFunc(func(ctx context.Context, input any) any {
					return map[string]any{
						"message": "should be overridden",
						"data":    input,
					}
				}),
				errorTransformer: nil,
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						responder.Respond(request.Context(), writer, nil, Option.Respond().Transformer(BodyTransformFunc(func(_ context.Context, input any) any {
							return map[string]any{
								"message": "custom body transform",
								"data":    input,
							}
						})))
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected content type to be 'application/json', but got '%s'", resp.Header.Get("Content-Type"))
				}
				if resp.Header.Get("Content-Encoding") != "" {
					t.Errorf("Expected content encoding to be empty, got %s", resp.Header.Get("Content-Encoding"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) == 0 {
					t.Errorf("Expected body to be not empty")
					return
				}
				log := `
				{
				  "time": "<<PRESENCE>>",
				  "message": "GET /",
				  "caller": "<<PRESENCE>>",
				  "level": "info",
				  "service": {
					"name": "responder-test",
					"environment": "testing",
					"type": "unit-test"
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
					  "status": %d,
                      "body": {
						"message": "custom body transform",
						"data": null
					  },
                      "headers": {
						"Content-Type": [
						  "application/json"
						],
						"Content-Length": [
						  "%s"
						]
					  }
					}
				  }
				}`
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host, resp.StatusCode, resp.Header.Get("Content-Length"))
			},
		},
		{
			name: "nil body - custom BodyTransform - gzip compression - skip on data too small",
			fields: fields{
				encoder: NewJSONEncoder(),
				transformer: BodyTransformFunc(func(ctx context.Context, input any) any {
					return map[string]any{
						"message": "gzipped body",
						"data":    input,
					}
				}),
				errorTransformer: nil,
				compressor:       NewGzipCompression(),
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						responder.Respond(request.Context(), writer, nil)
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected content type to be 'application/json', but got '%s'", resp.Header.Get("Content-Type"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) == 0 {
					t.Errorf("Expected body to be not empty")
					return
				}
				log := `
				{
					"time": "<<PRESENCE>>",
					"message": "GET /",
					"caller": "<<PRESENCE>>",
					"level": "info",
					"service": {
						"name": "responder-test",
						"environment": "testing",
						"type": "unit-test"
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
							"body": {
								"data": null,
								"message": "gzipped body"
							},
							"headers": {
								"Content-Length": [
									"%d"
								],
								"Content-Type": [
									"application/json"
								]
							},
							"status": 200
						}
					}
				}
				`
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host, len(body))
			},
		},
		{
			name: "nil body - custom BodyTransform - gzip compression",
			fields: fields{
				encoder: NewJSONEncoder(),
				transformer: BodyTransformFunc(func(ctx context.Context, input any) any {
					return map[string]any{
						"message": "gzipped body",
						"data":    input,
					}
				}),
				errorTransformer: nil,
				compressor:       NewGzipCompression(),
				callerDepth:      2,
			},
			gen: gen{
				server: func(responder *Responder, middleware Middleware) *httptest.Server {
					handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						input := strings.Repeat("foo ", 400)
						responder.Respond(request.Context(), writer, input)
					}))
					return httptest.NewServer(handler)
				},
				tower: towerGen,
			},
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected content type to be 'application/json', but got '%s'", resp.Header.Get("Content-Type"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %s", err.Error())
					return
				}
				if len(body) == 0 {
					t.Errorf("Expected body to be not empty")
					return
				}
				input := strings.Repeat("foo ", 400)
				log := `
				{
					"time": "<<PRESENCE>>",
					"message": "GET /",
					"caller": "<<PRESENCE>>",
					"level": "info",
					"service": {
						"name": "responder-test",
						"environment": "testing",
						"type": "unit-test"
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
							"body": {
								"data": "%s",
								"message": "gzipped body"
							},
							"headers": {
								"Content-Length": [
									"77"
								],
								"Content-Encoding": [ "gzip" ],
								"Content-Type": [
									"application/json"
								]
							},
							"status": 200
						}
					}
				}
				`
				j := jsonassert.New(t)
				j.Assertf(logger.String(), log, resp.Request.Host, input)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tower.NewTestingJSONLogger()
			tow := tt.gen.tower(logger)
			r := Responder{
				encoder:          tt.fields.encoder,
				transformer:      tt.fields.transformer,
				errorTransformer: tt.fields.errorTransformer,
				tower:            tow,
				compressor:       tt.fields.compressor,
				callerDepth:      tt.fields.callerDepth,
			}
			middleware := LoggingMiddleware(NewServerLogger())
			server := tt.gen.server(&r, middleware)
			defer server.Close()
			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			if err != nil {
				t.Errorf("Error creating request: %s", err.Error())
				return
			}
			if tt.fields.compressor.ContentEncoding() != "" {
				req.Header.Set("Accept-Encoding", tt.fields.compressor.ContentEncoding())
			}
			// req.Close prevents the client from reusing the connection
			req.Close = true
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			tt.test(t, resp, logger)
			err = resp.Body.Close()
			if err != nil {
				t.Fatalf("Error closing response body: %s", err.Error())
			}
			if t.Failed() {
				logger.PrettyPrint()
			}
		})
	}
}
