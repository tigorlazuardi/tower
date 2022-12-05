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
	if os.Getenv("TOWERHTTP_TEST_GLOBAL_RESPOND_OK") == "" {
		t.Skip("skipping test; set TOWERHTTP_TEST_GLOBAL_RESPOND_OK to run")
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
			name: "expected caller location",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					towerhttp.Respond(w, r, nil)
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
			towerhttp.Exported.Responder().SetTower(tow)
			towerhttp.Exported.Responder().RegisterHook(towerhttp.NewLoggerHook())
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
