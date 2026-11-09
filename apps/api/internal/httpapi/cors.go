package httpapi

import (
	"net/http"
	"os"
	"strings"
)

const defaultDevOrigin = "http://localhost:5173"

// allowedOriginsFromEnv читает CORS_ORIGINS через запятую.
// Пустое значение — дефолт для локальной разработки.
func allowedOriginsFromEnv() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if strings.TrimSpace(raw) == "" {
		return []string{defaultDevOrigin}
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if o := strings.TrimSpace(p); o != "" {
			out = append(out, o)
		}
	}
	if len(out) == 0 {
		return []string{defaultDevOrigin}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}

// corsMiddleware — preflight и CORS-заголовки для /api.
// Без Origin — пропускает как есть (curl, пробы, админка).
// Origin не из списка — 403 без CORS-заголовков.
func corsMiddleware(next http.Handler, allowed []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !originAllowed(origin, allowed) {
			writeError(w, http.StatusForbidden, "origin не разрешён")
			return
		}
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
