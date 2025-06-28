package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type IUserModel interface {
	Insert(user *User) error
	Update(user *User) error
	GetByToken(scope, token string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByID(id int64) (*User, error)
}

type ITokenModel interface {
	New(userID int64, ttl time.Duration, scope string) (*Token, error)
	Insert(token *Token) error
	DeleteTokensByUser(scope string, userID int64) error
}

type IBookModel interface {
	Delete(id int64) error
	Get(id int64) (*Book, error)
	GetAll(title string, filters Filters) ([]*Book, *Metadata, error)
	GetAllByUser(title string, filters Filters, userID int64) ([]*Book, *Metadata, error)
	Insert(book *Book) error
	Update(book *Book) error
}

type IChapterModel interface {
	Insert(chapter *Chapter) error
}

type Models struct {
	Users    IUserModel
	Tokens   ITokenModel
	Books    IBookModel
	Chapters IChapterModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:    UserModel{DB: db},
		Tokens:   TokenModel{DB: db},
		Books:    BookModel{DB: db},
		Chapters: ChapterModel{DB: db},
	}
}
