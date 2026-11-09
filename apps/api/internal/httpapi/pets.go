package httpapi

import (
	"net/http"
	"strings"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

type petRequest struct {
	Name      string `json:"name"`
	Species   string `json:"species"`
	Sex       string `json:"sex"`
	BirthYear *int   `json:"birth_year"`
	Breed     string `json:"breed"`
	Traits    string `json:"traits"`
}

func (s *Server) handleCreatePet(w http.ResponseWriter, r *http.Request) {
	var req petRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "тело запроса не разобрано: "+err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 120 {
		writeError(w, http.StatusBadRequest, "name питомца пустой или длиннее 120 символов")
		return
	}
	species := req.Species
	if species == "" {
		species = "dog"
	}
	sex := req.Sex
	if sex == "" {
		sex = "unknown"
	}
	if sex != "male" && sex != "female" && sex != "unknown" {
		writeError(w, http.StatusBadRequest, "sex принимает male, female или unknown")
		return
	}
	if req.BirthYear != nil && (*req.BirthYear < 1980 || *req.BirthYear > 2100) {
		writeError(w, http.StatusBadRequest, "birth_year вне разумного диапазона")
		return
	}

	pet, err := s.store.CreatePet(r.Context(), store.Pet{
		UserID:    userFrom(r).ID,
		Name:      name,
		Species:   species,
		Sex:       sex,
		BirthYear: req.BirthYear,
		Breed:     strings.TrimSpace(req.Breed),
		Traits:    strings.TrimSpace(req.Traits),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	writeJSON(w, http.StatusCreated, pet)
}

func (s *Server) handleListPets(w http.ResponseWriter, r *http.Request) {
	pets, err := s.store.ListPets(r.Context(), userFrom(r).ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка базы")
		return
	}
	if pets == nil {
		pets = []store.Pet{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": pets})
}
