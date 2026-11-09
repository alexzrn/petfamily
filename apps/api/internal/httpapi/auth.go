package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  userPublic `json:"user"`
}

type userPublic struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func publicUser(u store.User) userPublic {
	return userPublic{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName}
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "тело запроса не разобрано: "+err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.DisplayName)
	switch {
	case !strings.Contains(email, "@") || len(email) < 3:
		writeError(w, http.StatusBadRequest, "email не похож на адрес")
		return
	case len(req.Password) < 8:
		writeError(w, http.StatusBadRequest, "password короче 8 символов")
		return
	case name == "" || len(name) > 120:
		writeError(w, http.StatusBadRequest, "display_name пустой или длиннее 120 символов")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось захешировать пароль")
		return
	}
	user, err := s.store.CreateUser(r.Context(), email, string(hash), name)
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "такой email уже зарегистрирован")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}

	token, err := s.issueToken(w, r)
	if err != nil {
		return
	}
	if err := s.store.CreateSession(r.Context(), user.ID, token, s.sessionTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать сессию")
		return
	}
	writeJSON(w, http.StatusCreated, authResponse{Token: token, User: publicUser(user)})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "тело запроса не разобрано: "+err.Error())
		return
	}
	user, err := s.store.UserByEmail(r.Context(), req.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}
	token, err := s.issueToken(w, r)
	if err != nil {
		return
	}
	if err := s.store.CreateSession(r.Context(), user.ID, token, s.sessionTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать сессию")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: publicUser(user)})
}

func (s *Server) issueToken(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := newToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось выпустить токен")
		return "", err
	}
	return token, nil
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Context().Value(ctxToken).(string)
	if err := s.store.DeleteSession(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, publicUser(userFrom(r)))
}
