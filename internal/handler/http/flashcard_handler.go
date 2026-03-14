package http

import (
	"errors"
	"net/http"

	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/Boyeep/flippy-backend/internal/repository"
	"github.com/Boyeep/flippy-backend/internal/service"
)

type FlashcardHandler struct {
	service service.FlashcardService
}

func NewFlashcardHandler(service service.FlashcardService) FlashcardHandler {
	return FlashcardHandler{service: service}
}

func (h FlashcardHandler) ListBySet(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPublicBySetSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load flashcards")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h FlashcardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var input domain.CreateFlashcardInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Create(r.Context(), userID, r.PathValue("slug"), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid flashcard payload")
		case errors.Is(err, repository.ErrFlashcardSetNotFound):
			writeError(w, http.StatusNotFound, "flashcard set not found")
		case errors.Is(err, repository.ErrFlashcardConflict):
			writeError(w, http.StatusConflict, "flashcard position already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create flashcard")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h FlashcardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var input domain.UpdateFlashcardInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Update(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid flashcard payload")
		case errors.Is(err, repository.ErrFlashcardNotFound):
			writeError(w, http.StatusNotFound, "flashcard not found")
		case errors.Is(err, repository.ErrFlashcardConflict):
			writeError(w, http.StatusConflict, "flashcard position already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update flashcard")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h FlashcardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	if err := h.service.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		if errors.Is(err, repository.ErrFlashcardNotFound) {
			writeError(w, http.StatusNotFound, "flashcard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete flashcard")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
