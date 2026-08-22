package httptransport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/bootstrap"
	"github.com/wyw14/cry-081/internal/config"
	"go.uber.org/zap"
)

func TestRouterHealthAndStableValidationError(t *testing.T) {
	cfg := testConfig()
	app, err := bootstrap.Build(context.Background(), cfg, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	health := httptest.NewRecorder()
	app.Router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", health.Code, health.Body.String())
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"not-an-email"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "request-test-1")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("validation status=%d body=%s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "VALIDATION_FAILED" || body["request_id"] != "request-test-1" {
		t.Fatalf("unstable error response: %#v", body)
	}
}

func TestProtectedRouteRequiresActor(t *testing.T) {
	cfg := testConfig()
	app, _ := bootstrap.Build(context.Background(), cfg, zap.NewNop())
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/manuscripts", nil))
	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "AUTHENTICATION_REQUIRED") {
		t.Fatalf("protected route status=%d body=%s", response.Code, response.Body.String())
	}
}

func testConfig() config.Config {
	return config.Config{
		Environment: "test",
		HTTP: config.HTTP{
			Address: ":0", AllowedOrigin: "http://localhost:5173",
			RequestTimeout: time.Second, ShutdownTimeout: time.Second, RequestsPerMin: 100,
		},
		Tokens:  config.Tokens{AccessTTL: time.Minute, RefreshTTL: time.Hour},
		Storage: config.Storage{Driver: "memory"},
	}
}
