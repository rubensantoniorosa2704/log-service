package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log/dto"
	logCtrl "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/controller/log"
)

// --- stubs ---

type stubLogUsecase struct{}

func (s *stubLogUsecase) CreateLog(_ context.Context, _ dto.CreateLogInput) (*dto.CreateLogOutput, error) {
	return &dto.CreateLogOutput{}, nil
}

// stubSSEServer simulates the real SSE server: it sets Cache-Control: no-cache
// before writing the response, matching the native SSE implementation.
type stubSSEServer struct{}

func (s *stubSSEServer) Subscribe(w http.ResponseWriter, _ *http.Request, _ string) {
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
}

func newTestRouter() http.Handler {
	return RegisterRoutes(RouterConfig{
		LogController: logCtrl.NewLogController(&stubLogUsecase{}),
		SSEServer:     &stubSSEServer{},
	})
}

// --- helpers ---

func assertHeader(t *testing.T, h http.Header, key, want string) {
	t.Helper()
	if got := h.Get(key); got != want {
		t.Errorf("header %q: want %q, got %q", key, want, got)
	}
}

func assertHeaderNotEqual(t *testing.T, h http.Header, key, notWant string) {
	t.Helper()
	if got := h.Get(key); got == notWant {
		t.Errorf("header %q should not be %q", key, notWant)
	}
}

// --- tests ---

// TestSecurityHeaders_APIRoute validates that standard security headers are
// applied to a regular API route (ADR-001 §Headers Applied).
func TestSecurityHeaders_APIRoute(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	cases := []struct{ key, want string }{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Content-Security-Policy", "default-src 'none'"},
		{"Cache-Control", "no-store"},
	}
	for _, tc := range cases {
		assertHeader(t, w.Result().Header, tc.key, tc.want)
	}
}

// TestSecurityHeaders_SwaggerExclusion validates that /swagger/* and /docs/*
// routes do NOT receive the restrictive CSP (ADR-001 §Swagger / Docs exception).
func TestSecurityHeaders_SwaggerExclusion(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	for _, path := range []string{"/swagger/index.html", "/docs/swagger.json"} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assertHeaderNotEqual(t, w.Result().Header, "Content-Security-Policy", "default-src 'none'")
		})
	}
}

// TestSecurityHeaders_SwaggerHasOtherHeaders validates that non-CSP security
// headers are still present on Swagger routes.
func TestSecurityHeaders_SwaggerHasOtherHeaders(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertHeader(t, w.Result().Header, "X-Content-Type-Options", "nosniff")
	assertHeader(t, w.Result().Header, "X-Frame-Options", "DENY")
	assertHeader(t, w.Result().Header, "Referrer-Policy", "strict-origin-when-cross-origin")
}

// TestSecurityHeaders_SSEExclusion validates that the SSE endpoint does NOT
// receive Cache-Control: no-store and that the SSE library's own Cache-Control
// is preserved (ADR-001 §SSE exception).
func TestSecurityHeaders_SSEExclusion(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/app-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Global middleware must NOT overwrite SSE's Cache-Control.
	assertHeaderNotEqual(t, w.Result().Header, "Cache-Control", "no-store")
	// SSE library sets no-cache; verify it is preserved.
	assertHeader(t, w.Result().Header, "Cache-Control", "no-cache")
}

// TestSecurityHeaders_SSEHasOtherHeaders validates that non-Cache-Control
// security headers are still applied to the SSE route.
func TestSecurityHeaders_SSEHasOtherHeaders(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/app-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertHeader(t, w.Result().Header, "X-Content-Type-Options", "nosniff")
	assertHeader(t, w.Result().Header, "X-Frame-Options", "DENY")
	assertHeader(t, w.Result().Header, "Referrer-Policy", "strict-origin-when-cross-origin")
}

// TestCORS_SingleSource is a regression test that validates CORS headers come
// from the Chi middleware and are present on all routes including SSE
// (ADR-001 §CORS Consolidation).
func TestCORS_SingleSource(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	routes := []struct {
		method, path string
	}{
		{http.MethodPost, "/api/v1/logs"},
		{http.MethodGet, "/api/v1/events/app-123"},
	}

	for _, tc := range routes {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Origin", "http://example.com")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if got := w.Result().Header.Get("Access-Control-Allow-Origin"); got == "" {
				t.Errorf("path %q: expected Access-Control-Allow-Origin to be set by Chi CORS middleware", tc.path)
			}
		})
	}
}

// TestCORS_NoDuplication is an architectural regression test: the SSE server
// must NOT set its own Access-Control-* headers. If s.Headers is re-introduced
// in sse/server.go this test will catch duplicate values.
func TestCORS_NoDuplication(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/app-123", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// http.Header stores multiple values; len > 1 means duplication.
	values := w.Result().Header["Access-Control-Allow-Origin"]
	if len(values) > 1 {
		t.Errorf("Access-Control-Allow-Origin duplicated: %v — CORS must be set in one place only", values)
	}
}
