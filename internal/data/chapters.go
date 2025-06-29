package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

type Chapter struct {
	ID          int64     `json:"id"`
	ChapterNo   int64     `json:"chapterNo"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     *string   `json:"content"`
	BookID      int64     `json:"bookId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"-"`
}

type ChapterModel struct {
	DB *sql.DB
}

func ValidateChapter(v *validator.Validator, chapter *Chapter) {
	v.Check(chapter.Title != "", "title", "must be provided")
	v.Check(len(chapter.Title) <= 200, "title", "must not be more than 200 bytes long")

	v.Check(chapter.BookID > 0, "bookId", "must be provided")

	if chapter.Description != "" {
		v.Check(len(chapter.Description) <= 500, "description", "must not be more than 500 bytes long")
	}

}

func (m ChapterModel) Insert(chapter *Chapter) error {
	query := `
		INSERT INTO chapters (title,description,book_id)
		VALUES ($1,$2,$3)
		RETURNING id,created_at, version
	`

	args := []any{chapter.Title, chapter.Description, chapter.BookID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&chapter.ID, &chapter.CreatedAt, &chapter.Version)
}

func (m ChapterModel) GetByBookId(bookId int64) ([]*Chapter, error) {
	query := `
		SELECT id, created_at, updated_at, title, description, chapter_no, content, book_id, version
		FROM chapters 
		WHERE book_id = $1
	`

	args := []any{bookId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	chapters := []*Chapter{}

	for rows.Next() {
		var chapter Chapter
		err := rows.Scan(
			&chapter.ID,
			&chapter.CreatedAt,
			&chapter.UpdatedAt,
			&chapter.Title,
			&chapter.Description,
			&chapter.ChapterNo,
			&chapter.Content,
			&chapter.BookID,
			&chapter.Version,
		)

		if err != nil {
			return nil, err
		}

		chapters = append(chapters, &chapter)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chapters, nil

}
