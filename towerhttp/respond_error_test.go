package towerhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponder_RespondError(t *testing.T) {
	type fields struct {
		encoder          Encoder
		transformer      BodyTransformer
		errorTransformer ErrorBodyTransformer
		compressor       Compressor
		callerDepth      int
	}
	type args struct {
		ctx        context.Context
		rw         http.ResponseWriter
		errPayload error
		opts       []RespondOption
	}
	type testRequestGenerator = func(server *httptest.Server) *http.Request
	towerGen := func(logger tower.Logger) *tower.Tower {
		t := tower.NewTower(tower.Service{
			Name:        "responder-test",
			Environment: "testing",
			Type:        "unit-test",
		})
		t.SetLogger(logger)
		return t
	}
	//getRequest := func(server *httptest.Server) *http.Request {
	//	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	return req
	//}
	postRequest := func(body io.ReadCloser) testRequestGenerator {
		return func(server *httptest.Server) *http.Request {
			req, err := http.NewRequest(http.MethodPost, server.URL, body)
			if err != nil {
				t.Fatal(err)
			}
			return req
		}
	}
	mustJsonBody := func(body any) io.ReadCloser {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		return io.NopCloser(bytes.NewReader(b))
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		server  func(*Responder) *httptest.Server
		request func(server *httptest.Server) *http.Request
		test    func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger)
	}{
		{
			name: "common pattern",
			fields: fields{
				encoder:          NewJSONEncoder(),
				transformer:      NoopBodyTransform{},
				errorTransformer: SimpleErrorTransformer{},
				compressor:       NoCompression{},
				callerDepth:      2,
			},
			args: args{},
			server: func(responder *Responder) *httptest.Server {
				handler := responder.RequestBodyCloner()(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					_, err := io.ReadAll(request.Body)
					if err != nil {
						t.Fatalf("failed to read request body: %v", err)
					}
					responder.RespondError(writer, request, errors.New("test error"))
				}))
				return httptest.NewServer(handler)
			},
			request: postRequest(mustJsonBody(map[string]any{"foo": "bar"})),
			test: func(t *testing.T, resp *http.Response, logger *tower.TestingJSONLogger) {
				if resp.StatusCode != http.StatusInternalServerError {
					t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected content type %s, got %s", "application/json", resp.Header.Get("Content-Type"))
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}
				if len(body) == 0 {
					t.Error("expected response body, got empty")
				}
				wantBody := `{"error":"test error"}`
				j := jsonassert.New(t)
				j.Assertf(string(body), wantBody)
				wantLog := `
				{
					"time": "<<PRESENCE>>",
					"code": 500,
					"message": "test error",
					"caller": "<<PRESENCE>>",
					"level": "error",
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
							"method": "POST",
							"url": "%s/",
							"body": {"foo":"bar"}
						},
						"response": {
							"body": {
								"error": "test error"
							},
							"headers": {
								"Content-Length": [
									"23"
								],
								"Content-Type": [
									"application/json"
								]
							},
							"status": 500
						}
					},
					"error": {
						"summary": "test error"
					}
				}`
				j.Assertf(logger.String(), wantLog, resp.Request.Host)
				if !strings.Contains(logger.String(), "towerhttp/respond_error_test.go") {
					t.Error("expected caller to be in towerhttp/respond_error_test.go")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tower.NewTestingJSONLogger()
			tow := towerGen(logger)
			r := Responder{
				encoder:          tt.fields.encoder,
				transformer:      tt.fields.transformer,
				errorTransformer: tt.fields.errorTransformer,
				tower:            tow,
				compressor:       tt.fields.compressor,
				callerDepth:      tt.fields.callerDepth,
			}
			r.RegisterHook(NewLoggerHook())
			server := tt.server(&r)
			defer server.Close()
			resp, err := http.DefaultClient.Do(tt.request(server))
			if err != nil {
				t.Fatal(err)
			}
			tt.test(t, resp, logger)
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if t.Failed() {
				logger.PrettyPrint()
			}
		})
	}
}
