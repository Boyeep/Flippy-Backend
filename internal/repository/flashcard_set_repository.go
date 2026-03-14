package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrFlashcardSetNotFound = errors.New("flashcard set not found")
	ErrFlashcardSetConflict = errors.New("flashcard set already exists")
)

type FlashcardSetRepository struct {
	db *pgxpool.Pool
}

func NewFlashcardSetRepository(db *pgxpool.Pool) FlashcardSetRepository {
	return FlashcardSetRepository{db: db}
}

func (r FlashcardSetRepository) ListPublic(ctx context.Context) ([]domain.FlashcardSet, error) {
	query := `
		SELECT id, owner_id, course_id, slug, title, description, visibility, status, language_code, card_count, COALESCE(estimated_minutes, 0), created_at, updated_at
		FROM flashcard_sets
		WHERE visibility = 'public' AND status = 'published'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFlashcardSets(rows)
}

func (r FlashcardSetRepository) FindBySlug(ctx context.Context, slug string) (domain.FlashcardSet, error) {
	query := `
		SELECT id, owner_id, course_id, slug, title, description, visibility, status, language_code, card_count, COALESCE(estimated_minutes, 0), created_at, updated_at
		FROM flashcard_sets
		WHERE slug = $1
	`

	item, err := scanFlashcardSet(r.db.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FlashcardSet{}, ErrFlashcardSetNotFound
		}
		return domain.FlashcardSet{}, err
	}

	return item, nil
}

func (r FlashcardSetRepository) Create(ctx context.Context, ownerID, slug string, input domain.CreateFlashcardSetInput) (domain.FlashcardSet, error) {
	query := `
		INSERT INTO flashcard_sets (owner_id, course_id, slug, title, description, visibility, status, language_code, estimated_minutes)
		VALUES (
			$1,
			CASE WHEN $2 = '' THEN NULL ELSE $2::uuid END,
			$3,
			$4,
			NULLIF($5, ''),
			$6,
			$7,
			$8,
			NULLIF($9, 0)
		)
		RETURNING id, owner_id, course_id, slug, title, description, visibility, status, language_code, card_count, COALESCE(estimated_minutes, 0), created_at, updated_at
	`

	item, err := scanFlashcardSet(r.db.QueryRow(
		ctx,
		query,
		ownerID,
		strings.TrimSpace(input.CourseID),
		slug,
		input.Title,
		input.Description,
		input.Visibility,
		input.Status,
		input.LanguageCode,
		input.EstimatedMinutes,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return domain.FlashcardSet{}, ErrFlashcardSetConflict
		}
		return domain.FlashcardSet{}, err
	}

	return item, nil
}

func (r FlashcardSetRepository) Update(ctx context.Context, ownerID, slug string, input domain.UpdateFlashcardSetInput) (domain.FlashcardSet, error) {
	query := `
		UPDATE flashcard_sets
		SET
			course_id = COALESCE(CASE WHEN $3 = '' THEN NULL ELSE $3::uuid END, course_id),
			title = COALESCE(NULLIF($4, ''), title),
			description = COALESCE($5, description),
			visibility = COALESCE(NULLIF($6, ''), visibility),
			status = COALESCE(NULLIF($7, ''), status),
			language_code = COALESCE(NULLIF($8, ''), language_code),
			estimated_minutes = COALESCE($9, estimated_minutes),
			updated_at = NOW()
		WHERE owner_id = $1 AND slug = $2
		RETURNING id, owner_id, course_id, slug, title, description, visibility, status, language_code, card_count, COALESCE(estimated_minutes, 0), created_at, updated_at
	`

	item, err := scanFlashcardSet(r.db.QueryRow(
		ctx,
		query,
		ownerID,
		slug,
		stringValue(input.CourseID),
		stringValue(input.Title),
		input.Description,
		stringValue(input.Visibility),
		stringValue(input.Status),
		stringValue(input.LanguageCode),
		input.EstimatedMinutes,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FlashcardSet{}, ErrFlashcardSetNotFound
		}
		return domain.FlashcardSet{}, err
	}

	return item, nil
}

func (r FlashcardSetRepository) Delete(ctx context.Context, ownerID, slug string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM flashcard_sets WHERE owner_id = $1 AND slug = $2`, ownerID, slug)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrFlashcardSetNotFound
	}

	return nil
}

type flashcardSetScanner interface {
	Scan(dest ...any) error
}

func scanFlashcardSet(scanner flashcardSetScanner) (domain.FlashcardSet, error) {
	var item domain.FlashcardSet
	var courseID *string
	var description *string

	err := scanner.Scan(
		&item.ID,
		&item.OwnerID,
		&courseID,
		&item.Slug,
		&item.Title,
		&description,
		&item.Visibility,
		&item.Status,
		&item.LanguageCode,
		&item.CardCount,
		&item.EstimatedMinutes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.FlashcardSet{}, err
	}

	if courseID != nil {
		item.CourseID = *courseID
	}
	if description != nil {
		item.Description = *description
	}

	return item, nil
}

func scanFlashcardSets(rows pgx.Rows) ([]domain.FlashcardSet, error) {
	items := make([]domain.FlashcardSet, 0)
	for rows.Next() {
		item, err := scanFlashcardSet(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

func stringValue(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}
