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
	ErrFlashcardNotFound = errors.New("flashcard not found")
	ErrFlashcardConflict = errors.New("flashcard already exists")
)

type FlashcardRepository struct {
	db *pgxpool.Pool
}

func NewFlashcardRepository(db *pgxpool.Pool) FlashcardRepository {
	return FlashcardRepository{db: db}
}

func (r FlashcardRepository) ListPublicBySetSlug(ctx context.Context, setSlug string) ([]domain.Flashcard, error) {
	query := `
		SELECT f.id, f.flashcard_set_id, f.position, f.question, f.answer, f.explanation, f.hint, f.created_at, f.updated_at
		FROM flashcards f
		INNER JOIN flashcard_sets fs ON fs.id = f.flashcard_set_id
		WHERE fs.slug = $1 AND fs.visibility = 'public' AND fs.status = 'published'
		ORDER BY f.position ASC, f.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, strings.TrimSpace(setSlug))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFlashcards(rows)
}

func (r FlashcardRepository) Create(ctx context.Context, ownerID, setSlug string, input domain.CreateFlashcardInput) (domain.Flashcard, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Flashcard{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var flashcardSetID string
	if err := tx.QueryRow(ctx, `SELECT id FROM flashcard_sets WHERE owner_id = $1 AND slug = $2`, ownerID, strings.TrimSpace(setSlug)).Scan(&flashcardSetID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flashcard{}, ErrFlashcardSetNotFound
		}
		return domain.Flashcard{}, err
	}

	query := `
		INSERT INTO flashcards (flashcard_set_id, position, question, answer, explanation, hint)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''))
		RETURNING id, flashcard_set_id, position, question, answer, explanation, hint, created_at, updated_at
	`

	item, err := scanFlashcard(tx.QueryRow(
		ctx,
		query,
		flashcardSetID,
		input.Position,
		input.Question,
		input.Answer,
		input.Explanation,
		input.Hint,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Flashcard{}, ErrFlashcardConflict
		}
		return domain.Flashcard{}, err
	}

	if _, err := tx.Exec(ctx, `UPDATE flashcard_sets SET card_count = card_count + 1, updated_at = NOW() WHERE id = $1`, flashcardSetID); err != nil {
		return domain.Flashcard{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Flashcard{}, err
	}

	return item, nil
}

func (r FlashcardRepository) Update(ctx context.Context, ownerID, flashcardID string, input domain.UpdateFlashcardInput) (domain.Flashcard, error) {
	query := `
		UPDATE flashcards f
		SET
			position = COALESCE($3, f.position),
			question = COALESCE(NULLIF($4, ''), f.question),
			answer = COALESCE(NULLIF($5, ''), f.answer),
			explanation = COALESCE($6, f.explanation),
			hint = COALESCE($7, f.hint),
			updated_at = NOW()
		FROM flashcard_sets fs
		WHERE f.flashcard_set_id = fs.id AND fs.owner_id = $1 AND f.id = $2
		RETURNING f.id, f.flashcard_set_id, f.position, f.question, f.answer, f.explanation, f.hint, f.created_at, f.updated_at
	`

	item, err := scanFlashcard(r.db.QueryRow(
		ctx,
		query,
		ownerID,
		strings.TrimSpace(flashcardID),
		input.Position,
		stringValue(input.Question),
		stringValue(input.Answer),
		input.Explanation,
		input.Hint,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flashcard{}, ErrFlashcardNotFound
		}
		if isUniqueViolation(err) {
			return domain.Flashcard{}, ErrFlashcardConflict
		}
		return domain.Flashcard{}, err
	}

	return item, nil
}

func (r FlashcardRepository) Delete(ctx context.Context, ownerID, flashcardID string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var flashcardSetID string
	if err := tx.QueryRow(ctx, `
		DELETE FROM flashcards f
		USING flashcard_sets fs
		WHERE f.flashcard_set_id = fs.id AND fs.owner_id = $1 AND f.id = $2
		RETURNING f.flashcard_set_id
	`, ownerID, strings.TrimSpace(flashcardID)).Scan(&flashcardSetID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFlashcardNotFound
		}
		return err
	}

	if _, err := tx.Exec(ctx, `UPDATE flashcard_sets SET card_count = GREATEST(card_count - 1, 0), updated_at = NOW() WHERE id = $1`, flashcardSetID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

type flashcardScanner interface {
	Scan(dest ...any) error
}

func scanFlashcard(scanner flashcardScanner) (domain.Flashcard, error) {
	var item domain.Flashcard
	var explanation *string
	var hint *string

	err := scanner.Scan(
		&item.ID,
		&item.FlashcardSetID,
		&item.Position,
		&item.Question,
		&item.Answer,
		&explanation,
		&hint,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Flashcard{}, err
	}

	if explanation != nil {
		item.Explanation = *explanation
	}
	if hint != nil {
		item.Hint = *hint
	}

	return item, nil
}

func scanFlashcards(rows pgx.Rows) ([]domain.Flashcard, error) {
	items := make([]domain.Flashcard, 0)
	for rows.Next() {
		item, err := scanFlashcard(rows)
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
