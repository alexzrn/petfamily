package httpapi

import (
	"net/http"
	"testing"
)

func TestRegisterRejectsShortPassword(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)

	code, body := c.do(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": "a@b.c", "password": "123", "display_name": "Леша",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("код %d, ожидали 400 (тело %v)", code, body)
	}
	msg, _ := body["error"].(string)
	if msg == "" {
		t.Fatalf("нет причины отказа в теле: %v", body)
	}
}

func TestRegisterLoginAndMe(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("owner@example.com")

	// регистрация выдала рабочий токен
	code, body := c.do(http.MethodGet, "/api/v1/me", nil)
	if code != http.StatusOK {
		t.Fatalf("/me: код %d, тело %v", code, body)
	}
	if body["email"] != "owner@example.com" {
		t.Fatalf("/me вернул не того: %v", body)
	}

	// вход по паролю выдаёт новый токен
	anon := newClient(t, srv)
	code, body = anon.do(http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "owner@example.com", "password": "правильный-пароль",
	})
	if code != http.StatusOK {
		t.Fatalf("логин: код %d, тело %v", code, body)
	}
	loginToken, _ := body["token"].(string)
	if loginToken == "" || loginToken == c.token {
		t.Fatalf("логин должен выдать новый токен, получили %q при старом %q", loginToken, c.token)
	}

	// неверный пароль
	code, body = anon.do(http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "owner@example.com", "password": "неверный-пароль",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("неверный пароль: код %d, тело %v", code, body)
	}

	// без токена доступа нет
	naked := newClient(t, srv)
	code, _ = naked.do(http.MethodGet, "/api/v1/me", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("без токена код %d, ожидали 401", code)
	}
}

func TestLogoutKillsSession(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("bye@example.com")

	code, _ := c.do(http.MethodPost, "/api/v1/auth/logout", nil)
	if code != http.StatusNoContent {
		t.Fatalf("выход: код %d, ожидали 204", code)
	}
	code, _ = c.do(http.MethodGet, "/api/v1/me", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("после выхода /me вернул %d, ожидали 401", code)
	}
}

func TestPets(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("pets@example.com")

	code, body := c.do(http.MethodPost, "/api/v1/pets", map[string]any{
		"name": "Тай", "sex": "male", "breed": "тайский риджбек", "birth_year": 2025,
	})
	if code != http.StatusCreated {
		t.Fatalf("создание питомца: код %d, тело %v", code, body)
	}

	code, body = c.do(http.MethodGet, "/api/v1/pets", nil)
	if code != http.StatusOK {
		t.Fatalf("список питомцев: код %d", code)
	}
	items, _ := body["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("питомцев %d, ожидали 1: %v", len(items), body)
	}

	code, body = c.do(http.MethodPost, "/api/v1/pets", map[string]any{"name": ""})
	if code != http.StatusBadRequest {
		t.Fatalf("пустое имя: код %d, ожидали 400 (тело %v)", code, body)
	}
}
