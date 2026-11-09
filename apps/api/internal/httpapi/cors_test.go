package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тесты CORS идут через настоящий Handler, но без базы:
// preflight и кривые запросы до store не доходят.
func testHandlerNoDB(t *testing.T) http.Handler {
	t.Helper()
	return NewServer(nil).Handler()
}

func corsReq(t *testing.T, h http.Handler, method, path, origin string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

const (
	goodOrigin = "https://petfamily.beliberde.net"
	evilOrigin = "https://evil.example"
)

func TestPreflightAllowed(t *testing.T) {
	t.Setenv("CORS_ORIGINS", goodOrigin)
	rr := corsReq(t, testHandlerNoDB(t), http.MethodOptions, "/api/v1/auth/login", goodOrigin)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("код %d, ожидали 204", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != goodOrigin {
		t.Fatalf("Allow-Origin %q, ожидали %q", got, goodOrigin)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "GET,POST,PATCH,DELETE,OPTIONS" {
		t.Fatalf("Allow-Methods %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
		t.Fatalf("Allow-Headers %q", got)
	}
	if got := rr.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("Max-Age %q", got)
	}
	if got := rr.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary %q", got)
	}
}

func TestPreflightDenied(t *testing.T) {
	t.Setenv("CORS_ORIGINS", goodOrigin)
	rr := corsReq(t, testHandlerNoDB(t), http.MethodOptions, "/api/v1/auth/login", evilOrigin)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("код %d, ожидали 403", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("чужому origin отдали Allow-Origin %q", got)
	}
}

func TestNormalAllowedOrigin(t *testing.T) {
	t.Setenv("CORS_ORIGINS", goodOrigin)
	// Пустое тело — 400 от валидации, но уже за CORS-слоем: заголовок должен быть.
	rr := corsReq(t, testHandlerNoDB(t), http.MethodPost, "/api/v1/auth/register", goodOrigin)
	if rr.Code == http.StatusForbidden {
		t.Fatalf("разрешённый origin получил 403")
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != goodOrigin {
		t.Fatalf("Allow-Origin %q, ожидали %q", got, goodOrigin)
	}
}

func TestNormalDeniedOrigin(t *testing.T) {
	t.Setenv("CORS_ORIGINS", goodOrigin)
	rr := corsReq(t, testHandlerNoDB(t), http.MethodPost, "/api/v1/auth/register", evilOrigin)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("код %d, ожидали 403", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("чужому origin отдали Allow-Origin %q", got)
	}
}

func TestNoOriginPassThrough(t *testing.T) {
	t.Setenv("CORS_ORIGINS", goodOrigin)
	// curl без Origin: никакой CORS-оценки, запрос идёт дальше (400 от валидации тела).
	rr := corsReq(t, testHandlerNoDB(t), http.MethodPost, "/api/v1/auth/register", "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("код %d, ожидали 400 от валидации (сквозной проход)", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("без Origin выставили Allow-Origin %q", got)
	}
}

func TestDefaultOrigin(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	allowed := allowedOriginsFromEnv()
	if len(allowed) != 1 || allowed[0] != "http://localhost:5173" {
		t.Fatalf("дефолт %v, ожидали [http://localhost:5173]", allowed)
	}
}
