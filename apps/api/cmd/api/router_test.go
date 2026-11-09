package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	newRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("код %d, ожидали 200", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Fatalf("тело %q, ожидали ok", rr.Body.String())
	}
}

func TestUnknownPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rr := httptest.NewRecorder()

	newRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("код %d, ожидали 404", rr.Code)
	}
}
