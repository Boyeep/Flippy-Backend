package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/Boyeep/flippy-backend/internal/repository"
)

type FlashcardService struct {
	repository repository.FlashcardRepository
}

func NewFlashcardService(repository repository.FlashcardRepository) FlashcardService {
	return FlashcardService{repository: repository}
}

func (s FlashcardService) ListPublicBySetSlug(ctx context.Context, setSlug string) ([]domain.Flashcard, error) {
	return s.repository.ListPublicBySetSlug(ctx, strings.TrimSpace(setSlug))
}

func (s FlashcardService) Create(ctx context.Context, ownerID, setSlug string, input domain.CreateFlashcardInput) (domain.Flashcard, error) {
	input = normalizeCreateFlashcardInput(input)
	if input.Position < 1 || input.Question == "" || input.Answer == "" {
		return domain.Flashcard{}, ErrInvalidInput
	}

	item, err := s.repository.Create(ctx, ownerID, strings.TrimSpace(setSlug), input)
	if err != nil {
		if errors.Is(err, repository.ErrFlashcardSetNotFound) || errors.Is(err, repository.ErrFlashcardConflict) {
			return domain.Flashcard{}, err
		}
		return domain.Flashcard{}, err
	}

	return item, nil
}

func (s FlashcardService) Update(ctx context.Context, ownerID, flashcardID string, input domain.UpdateFlashcardInput) (domain.Flashcard, error) {
	input = normalizeUpdateFlashcardInput(input)
	if input.Position != nil && *input.Position < 1 {
		return domain.Flashcard{}, ErrInvalidInput
	}
	if input.Question != nil && *input.Question == "" {
		return domain.Flashcard{}, ErrInvalidInput
	}
	if input.Answer != nil && *input.Answer == "" {
		return domain.Flashcard{}, ErrInvalidInput
	}

	return s.repository.Update(ctx, ownerID, strings.TrimSpace(flashcardID), input)
}

func (s FlashcardService) Delete(ctx context.Context, ownerID, flashcardID string) error {
	return s.repository.Delete(ctx, ownerID, strings.TrimSpace(flashcardID))
}

func normalizeCreateFlashcardInput(input domain.CreateFlashcardInput) domain.CreateFlashcardInput {
	input.Question = strings.TrimSpace(input.Question)
	input.Answer = strings.TrimSpace(input.Answer)
	input.Explanation = strings.TrimSpace(input.Explanation)
	input.Hint = strings.TrimSpace(input.Hint)
	return input
}

func normalizeUpdateFlashcardInput(input domain.UpdateFlashcardInput) domain.UpdateFlashcardInput {
	if input.Question != nil {
		value := strings.TrimSpace(*input.Question)
		input.Question = &value
	}
	if input.Answer != nil {
		value := strings.TrimSpace(*input.Answer)
		input.Answer = &value
	}
	if input.Explanation != nil {
		value := strings.TrimSpace(*input.Explanation)
		input.Explanation = &value
	}
	if input.Hint != nil {
		value := strings.TrimSpace(*input.Hint)
		input.Hint = &value
	}
	return input
}
