package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

type meetupRequest struct {
	Title    string  `json:"title"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	StartsAt string  `json:"starts_at"`
}

func (s *Server) handleCreateMeetup(w http.ResponseWriter, r *http.Request) {
	var req meetupRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "тело запроса не разобрано: "+err.Error())
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > 160 {
		writeError(w, http.StatusBadRequest, "title пустой или длиннее 160 символов")
		return
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		writeError(w, http.StatusBadRequest, "координаты встречи вне допустимых значений")
		return
	}
	starts, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "starts_at не в формате RFC3339")
		return
	}
	if starts.Before(time.Now().Add(-time.Minute)) {
		writeError(w, http.StatusBadRequest, "starts_at уже в прошлом")
		return
	}

	m, err := s.store.CreateMeetup(r.Context(), store.Meetup{
		UserID:   userFrom(r).ID,
		Title:    title,
		Lat:      req.Lat,
		Lng:      req.Lng,
		StartsAt: starts,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сохранить встречу")
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (s *Server) handleListMeetups(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
	lng, err2 := strconv.ParseFloat(q.Get("lng"), 64)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "нужны числовые lat и lng")
		return
	}
	radius := 3000.0
	if raw := q.Get("radius_m"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v <= 0 || v > 100000 {
			writeError(w, http.StatusBadRequest, "radius_m от 1 до 100000")
			return
		}
		radius = v
	}
	limit := 50
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 200 {
			writeError(w, http.StatusBadRequest, "limit от 1 до 200")
			return
		}
		limit = n
	}

	items, err := s.store.ListMeetups(r.Context(), lat, lng, radius, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	if items == nil {
		items = []store.Meetup{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
