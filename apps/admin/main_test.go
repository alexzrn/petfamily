package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testApp() *app {
	return &app{tmpl: parseTemplates(), user: "admin", password: "секрет"}
}

func TestHealthzOpen(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	testApp().routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz: код %d", rr.Code)
	}
}

func TestGuardRejectsAnonymous(t *testing.T) {
	for _, path := range []string{"/", "/users", "/walks", "/meetups"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		testApp().routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s: код %d, ожидали 401", path, rr.Code)
		}
		if rr.Header().Get("WWW-Authenticate") == "" {
			t.Fatalf("%s: нет заголовка WWW-Authenticate", path)
		}
	}
}

func TestGuardRejectsWrongPassword(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.SetBasicAuth("admin", "не тот")
	rr := httptest.NewRecorder()
	testApp().routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("неверный пароль: код %d, ожидали 401", rr.Code)
	}
}

func TestFormatID(t *testing.T) {
	raw := make([]byte, 16)
	for i := range raw {
		raw[i] = byte(i)
	}
	got := formatID(raw)
	want := "00010203-0405-0607-0809-0a0b0c0d0e0f"
	if got != want {
		t.Fatalf("получили %q, ожидали %q", got, want)
	}
	if formatID([]byte{1, 2, 3}) != "" {
		t.Fatal("битый id должен давать пустую строку")
	}
}

func TestTemplatesRender(t *testing.T) {
	a := testApp()
	for _, name := range []string{"index", "users", "walks", "meetups"} {
		var sb strings.Builder
		if err := a.tmpl[name].ExecuteTemplate(&sb, "layout", map[string]any{}); err != nil {
			t.Fatalf("шаблон %s: %v", name, err)
		}
		if !strings.Contains(sb.String(), "PetFamily") {
			t.Fatalf("шаблон %s отрендерился без каркаса", name)
		}
	}
}
