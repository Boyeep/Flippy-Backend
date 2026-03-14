package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/Boyeep/flippy-backend/internal/repository"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

type FlashcardSetService struct {
	repository repository.FlashcardSetRepository
}

func NewFlashcardSetService(repository repository.FlashcardSetRepository) FlashcardSetService {
	return FlashcardSetService{repository: repository}
}

func (s FlashcardSetService) ListPublic(ctx context.Context) ([]domain.FlashcardSet, error) {
	return s.repository.ListPublic(ctx)
}

func (s FlashcardSetService) GetBySlug(ctx context.Context, slug string) (domain.FlashcardSet, error) {
	return s.repository.FindBySlug(ctx, strings.TrimSpace(slug))
}

func (s FlashcardSetService) Create(ctx context.Context, ownerID string, input domain.CreateFlashcardSetInput) (domain.FlashcardSet, error) {
	input = normalizeCreateInput(input)
	if input.Title == "" || !isValidVisibility(input.Visibility) || !isValidStatus(input.Status) {
		return domain.FlashcardSet{}, ErrInvalidInput
	}

	slug := buildSlug(input.Title)
	if slug == "" {
		return domain.FlashcardSet{}, ErrInvalidInput
	}

	item, err := s.repository.Create(ctx, ownerID, fmt.Sprintf("%s-%d", slug, time.Now().Unix()), input)
	if err != nil {
		if errors.Is(err, repository.ErrFlashcardSetConflict) {
			return domain.FlashcardSet{}, err
		}
		return domain.FlashcardSet{}, err
	}

	return item, nil
}

func (s FlashcardSetService) Update(ctx context.Context, ownerID, slug string, input domain.UpdateFlashcardSetInput) (domain.FlashcardSet, error) {
	input = normalizeUpdateInput(input)
	if input.Visibility != nil && !isValidVisibility(*input.Visibility) {
		return domain.FlashcardSet{}, ErrInvalidInput
	}
	if input.Status != nil && !isValidStatus(*input.Status) {
		return domain.FlashcardSet{}, ErrInvalidInput
	}

	return s.repository.Update(ctx, ownerID, strings.TrimSpace(slug), input)
}

func (s FlashcardSetService) Delete(ctx context.Context, ownerID, slug string) error {
	return s.repository.Delete(ctx, ownerID, strings.TrimSpace(slug))
}

func normalizeCreateInput(input domain.CreateFlashcardSetInput) domain.CreateFlashcardSetInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.CourseID = strings.TrimSpace(input.CourseID)
	input.Visibility = strings.TrimSpace(strings.ToLower(defaultIfEmpty(input.Visibility, "private")))
	input.Status = strings.TrimSpace(strings.ToLower(defaultIfEmpty(input.Status, "draft")))
	input.LanguageCode = strings.TrimSpace(strings.ToLower(defaultIfEmpty(input.LanguageCode, "id")))
	return input
}

func normalizeUpdateInput(input domain.UpdateFlashcardSetInput) domain.UpdateFlashcardSetInput {
	if input.CourseID != nil {
		value := strings.TrimSpace(*input.CourseID)
		input.CourseID = &value
	}
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		input.Title = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		input.Description = &value
	}
	if input.Visibility != nil {
		value := strings.TrimSpace(strings.ToLower(*input.Visibility))
		input.Visibility = &value
	}
	if input.Status != nil {
		value := strings.TrimSpace(strings.ToLower(*input.Status))
		input.Status = &value
	}
	if input.LanguageCode != nil {
		value := strings.TrimSpace(strings.ToLower(*input.LanguageCode))
		input.LanguageCode = &value
	}
	return input
}

func buildSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = slugSanitizer.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func isValidVisibility(value string) bool {
	switch value {
	case "private", "unlisted", "public":
		return true
	default:
		return false
	}
}

func isValidStatus(value string) bool {
	switch value {
	case "draft", "published", "archived":
		return true
	default:
		return false
	}
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
