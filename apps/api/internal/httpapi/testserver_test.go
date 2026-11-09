package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
	"github.com/alexzrn/petfamily/apps/api/internal/testsupport"
)

// newTestServer поднимает настоящий HTTP-сервер поверх настоящей MySQL.
// Без MYSQL_DSN тесты, требующие базу, пропускаются.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN не задан")
	}
	db, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("база: %v", err)
	}
	if err := testsupport.Reset(db); err != nil {
		t.Fatalf("подготовка базы: %v", err)
	}
	srv := httptest.NewServer(NewServer(store.NewStore(db)).Handler())
	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return srv
}

type client struct {
	t     *testing.T
	base  string
	token string
}

func newClient(t *testing.T, srv *httptest.Server) *client {
	return &client{t: t, base: srv.URL}
}

func (c *client) do(method, path string, body any) (int, map[string]any) {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.base+path, rdr)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out
}

func (c *client) register(email string) {
	c.t.Helper()
	code, body := c.do(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":        email,
		"password":     "правильный-пароль",
		"display_name": "Леша",
	})
	if code != http.StatusCreated {
		c.t.Fatalf("регистрация: код %d, тело %v", code, body)
	}
	c.token, _ = body["token"].(string)
	if c.token == "" {
		c.t.Fatal("регистрация не вернула токен")
	}
}
