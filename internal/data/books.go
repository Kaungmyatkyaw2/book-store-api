package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

type Book struct {
	ID           int64     `json:"id"`
	Title        string    `json:"string"`
	UserID       int64     `json:"userId"`
	CoverPicture string    `json:"coverPicture"`
	CreatedAt    time.Time `json:"createdAt"`
	Version      int       `json:"-"`
}

func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Title) <= 500, "title", "must not be more than 500 bytes long")

}

type BookModel struct {
	DB *sql.DB
}

func (m BookModel) GetAll() ([]*Book, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id,created_at,title,cover_picture,user_id,version 
	FROM books	
	`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
		)

		if err != nil {
			return nil, err
		}

		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil

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
