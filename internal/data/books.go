package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

type Book struct {
	ID           int64      `json:"id"`
	Title        string     `json:"string"`
	UserID       int64      `json:"userId"`
	CoverPicture string     `json:"coverPicture"`
	CreatedAt    time.Time  `json:"createdAt"`
	IsPublished  bool       `json:"isPublished"`
	PublishedAt  *time.Time `json:"publishedAt"`
	Version      int        `json:"-"`
}

func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Title) <= 500, "title", "must not be more than 500 bytes long")

}

type BookModel struct {
	DB *sql.DB
}

func getAllBooks(m BookModel, title string, filters Filters, userID int64) ([]*Book, *Metadata, error) {
	query := `
	SELECT count(*) OVER(), id,created_at,title,cover_picture,user_id,version, is_published, published_at 
	FROM books	
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple',$1) OR $1 = '') AND is_published = true
	`

	args := []any{title}

	if userID != -1 {
		query += ` AND user_id = $2`
		args = append(args, userID)
	}

	argsLength := len(args)

	query += fmt.Sprintf(`
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d
	`, filters.sortColumn(), filters.sortDirecton(), argsLength+1, argsLength+2)

	args = append(args, filters.limit())
	args = append(args, filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	totalRecords := 0
	books := []*Book{}

	for rows.Next() {
		var book Book

		err := rows.Scan(
			&totalRecords,
			&book.ID,
			&book.CreatedAt,
			&book.Title,
			&book.CoverPicture,
			&book.UserID,
			&book.Version,
			&book.IsPublished,
			&book.PublishedAt,
		)

		if err != nil {
			return nil, nil, err
		}

		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return books, &metadata, nil
}

func (m BookModel) GetAll(title string, filters Filters) ([]*Book, *Metadata, error) {
	return getAllBooks(m, title, filters, -1)
}

func (m BookModel) GetAllByUser(title string, filters Filters, userID int64) ([]*Book, *Metadata, error) {
	return getAllBooks(m, title, filters, userID)
}

func (m BookModel) Get(id int64) (*Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, cover_picture, version, user_id, is_published, published_at
		FROM books
		WHERE id = $1
	`

	var book Book

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&book.ID, &book.CreatedAt, &book.Title, &book.CoverPicture, &book.Version, &book.UserID, &book.IsPublished, &book.PublishedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &book, nil
}

func (m BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title,user_id,cover_picture)
		VALUES ($1,$2,$3)
		RETURNING id,created_at, version
	`

	args := []any{book.Title, book.UserID, book.CoverPicture}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.CreatedAt, &book.Version)
}

func (m BookModel) Update(book *Book) error {
	query := `
		UPDATE books 
		SET title = $1, cover_picture = $2, is_published = $3, published_at = $4
		WHERE id = $5 AND version = $6
		RETURNING version
	`

	args := []interface{}{
		book.Title,
		book.CoverPicture,
		book.IsPublished,
		book.PublishedAt,
		book.ID,
		book.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books 
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return nil
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}
