package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

type Server struct {
	store      *store.Store
	sessionTTL time.Duration
}

func NewServer(st *store.Store) *Server {
	return &Server{store: st, sessionTTL: 30 * 24 * time.Hour}
}

// Handler собирает все маршруты API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", s.auth(s.handleLogout))
	mux.HandleFunc("GET /api/v1/me", s.auth(s.handleMe))

	mux.HandleFunc("POST /api/v1/pets", s.auth(s.handleCreatePet))
	mux.HandleFunc("GET /api/v1/pets", s.auth(s.handleListPets))

	mux.HandleFunc("POST /api/v1/walks", s.auth(s.handleCreateWalk))
	mux.HandleFunc("GET /api/v1/walks", s.auth(s.handleListWalks))
	mux.HandleFunc("GET /api/v1/walks/{id}", s.auth(s.handleGetWalk))
	mux.HandleFunc("GET /api/v1/live/walks", s.auth(s.handleLiveWalks))

	mux.HandleFunc("POST /api/v1/meetups", s.auth(s.handleCreateMeetup))
	mux.HandleFunc("GET /api/v1/meetups", s.auth(s.handleListMeetups))

	return corsMiddleware(mux, allowedOriginsFromEnv())
}

type ctxKey string

const (
	ctxUser  ctxKey = "user"
	ctxToken ctxKey = "token"
)

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || len(token) < 32 {
			writeError(w, http.StatusUnauthorized, "нужен заголовок Authorization: Bearer <token>")
			return
		}
		user, err := s.store.UserByToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "сессия не найдена или истекла")
				return
			}
			writeError(w, http.StatusInternalServerError, "ошибка базы")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUser, user)
		ctx = context.WithValue(ctx, ctxToken, token)
		next(w, r.WithContext(ctx))
	}
}

func userFrom(r *http.Request) store.User {
	u, _ := r.Context().Value(ctxUser).(store.User)
	return u
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
