package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

type pointJSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	At  string  `json:"at"`
}

type walkRequest struct {
	PetID     *string     `json:"pet_id"`
	StartedAt string      `json:"started_at"`
	EndedAt   string      `json:"ended_at"`
	Points    []pointJSON `json:"points"`
}

const (
	maxWalkDuration = 12 * time.Hour
	maxWalkPoints   = 5000
)

func (s *Server) handleCreateWalk(w http.ResponseWriter, r *http.Request) {
	var req walkRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "тело запроса не разобрано: "+err.Error())
		return
	}
	started, err := time.Parse(time.RFC3339, req.StartedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "started_at не в формате RFC3339")
		return
	}
	ended, err := time.Parse(time.RFC3339, req.EndedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ended_at не в формате RFC3339")
		return
	}
	if !ended.After(started) {
		writeError(w, http.StatusBadRequest, "ended_at должен быть позже started_at")
		return
	}
	if ended.Sub(started) > maxWalkDuration {
		writeError(w, http.StatusBadRequest, "прогулка длиннее 12 часов не принимается")
		return
	}
	if len(req.Points) < 2 {
		writeError(w, http.StatusBadRequest, "в прогулке должно быть минимум две точки")
		return
	}
	if len(req.Points) > maxWalkPoints {
		writeError(w, http.StatusBadRequest, "слишком много точек в одной прогулке")
		return
	}

	points := make([]store.Point, 0, len(req.Points))
	for _, p := range req.Points {
		if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
			writeError(w, http.StatusBadRequest, "координаты точки вне допустимых значений")
			return
		}
		at, err := time.Parse(time.RFC3339, p.At)
		if err != nil {
			writeError(w, http.StatusBadRequest, "время точки не в формате RFC3339")
			return
		}
		points = append(points, store.Point{Lat: p.Lat, Lng: p.Lng, AtUTC: at})
	}

	user := userFrom(r)
	if req.PetID != nil {
		if _, err := s.store.PetOwned(r.Context(), user.ID, *req.PetID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusBadRequest, "питомец не найден среди своих")
				return
			}
			writeError(w, http.StatusInternalServerError, "ошибка базы")
			return
		}
	}

	walk, err := s.store.CreateWalk(r.Context(), store.NewWalk{
		UserID:    user.ID,
		PetID:     req.PetID,
		StartedAt: started,
		EndedAt:   ended,
		Points:    points,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сохранить прогулку")
		return
	}
	writeJSON(w, http.StatusCreated, walk)
}

func (s *Server) handleListWalks(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			writeError(w, http.StatusBadRequest, "limit от 1 до 100")
			return
		}
		limit = n
	}
	walks, err := s.store.ListWalks(r.Context(), userFrom(r).ID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	if walks == nil {
		walks = []store.Walk{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": walks})
}

func (s *Server) handleGetWalk(w http.ResponseWriter, r *http.Request) {
	walk, err := s.store.WalkByID(r.Context(), userFrom(r).ID, r.PathValue("id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// чужую прогулку отдаём как отсутствующую: не светим, что она есть
			writeError(w, http.StatusNotFound, "прогулка не найдена")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	writeJSON(w, http.StatusOK, walk)
}

func (s *Server) handleLiveWalks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
	lng, err2 := strconv.ParseFloat(q.Get("lng"), 64)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "нужны числовые lat и lng")
		return
	}
	meters := 2000.0
	if raw := q.Get("meters"); raw != "" {
		m, err := strconv.ParseFloat(raw, 64)
		if err != nil || m <= 0 || m > 50000 {
			writeError(w, http.StatusBadRequest, "meters от 1 до 50000")
			return
		}
		meters = m
	}
	items, err := s.store.LiveWalks(r.Context(), lat, lng, meters, 20*time.Minute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	if items == nil {
		items = []store.LiveWalk{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
