package towerhttp_test

import (
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/towerhttp"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGlobalRespond(t *testing.T) {
	const envKey = "TOWER_HTTP_TEST_EXPORTED"
	if os.Getenv(envKey) == "" {
		t.Skipf("skipping test; set %s env to run", envKey)
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
		server func() *httptest.Server
		test   func(t *testing.T, logger *tower.TestingJSONLogger)
	}{
		{
			name: "expected caller location for respond",
			server: func() *httptest.Server {
				handler := towerhttp.RequestBodyCloner()(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					towerhttp.Respond(writer, request, nil)
				}))
				return httptest.NewServer(handler)
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger) {
				if !strings.Contains(logger.String(), "towerhttp/respond_exported_test.go") {
					t.Error("expected caller location is correct")
				}
			},
		},
		{
			name: "expected caller location for respond error",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					towerhttp.RespondError(w, r, nil)
				}))
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger) {
				if !strings.Contains(logger.String(), "towerhttp/respond_exported_test.go") {
					t.Error("expected caller location is correct")
				}
			},
		},
		{
			name: "expected caller location for respond stream",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					towerhttp.RespondStream(w, r, "", nil)
				}))
			},
			test: func(t *testing.T, logger *tower.TestingJSONLogger) {
				if !strings.Contains(logger.String(), "towerhttp/respond_exported_test.go") {
					t.Error("expected caller location is correct")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tower.NewTestingJSONLogger()
			tow := towerGen(logger)
			r := towerhttp.NewResponder()
			r.SetTower(tow)
			r.RegisterHook(towerhttp.NewLoggerHook())
			r.SetCallerDepth(3)
			towerhttp.Exported.Responder().SetTower(tow)
			towerhttp.Exported.Responder().RegisterHook(towerhttp.NewLoggerHook())
			towerhttp.Exported.SetResponder(r)
			server := tt.server()
			defer server.Close()
			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			if err != nil {
				t.Errorf("Error creating request: %s", err.Error())
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			tt.test(t, logger)
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
