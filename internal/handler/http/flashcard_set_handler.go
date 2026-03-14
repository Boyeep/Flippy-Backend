package http

import (
	"errors"
	"net/http"

	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/Boyeep/flippy-backend/internal/repository"
	"github.com/Boyeep/flippy-backend/internal/service"
)

type FlashcardSetHandler struct {
	service service.FlashcardSetService
}

func NewFlashcardSetHandler(service service.FlashcardSetService) FlashcardSetHandler {
	return FlashcardSetHandler{service: service}
}

func (h FlashcardSetHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPublic(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load flashcard sets")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h FlashcardSetHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrFlashcardSetNotFound) {
			writeError(w, http.StatusNotFound, "flashcard set not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load flashcard set")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h FlashcardSetHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var input domain.CreateFlashcardSetInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid flashcard set payload")
		case errors.Is(err, repository.ErrFlashcardSetConflict):
			writeError(w, http.StatusConflict, "flashcard set already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create flashcard set")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h FlashcardSetHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var input domain.UpdateFlashcardSetInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Update(r.Context(), userID, r.PathValue("slug"), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid flashcard set payload")
		case errors.Is(err, repository.ErrFlashcardSetNotFound):
			writeError(w, http.StatusNotFound, "flashcard set not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update flashcard set")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h FlashcardSetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	if err := h.service.Delete(r.Context(), userID, r.PathValue("slug")); err != nil {
		if errors.Is(err, repository.ErrFlashcardSetNotFound) {
			writeError(w, http.StatusNotFound, "flashcard set not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete flashcard set")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
