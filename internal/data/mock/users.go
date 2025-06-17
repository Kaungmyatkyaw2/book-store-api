package mock

import (
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
)

var ValidActivationToken = "valid-token"

var MockUser = &data.User{
	ID:           1,
	Name:         "Alice",
	Email:        "alice@example.com",
	CreatedAt:    time.Date(2025, time.April, 3, 2, 2, 2, 2, time.UTC),
	Activated:    false,
	AuthProvider: data.CredentialAuthProvider,
}

type UserModel struct {
}

func (m *UserModel) Insert(user *data.User) error {

	if user.Email == "alice@example.com" {
		return data.ErrDuplicateEmail
	}

	return nil
}
func (m *UserModel) Update(user *data.User) error {

	if user.Email != "alice@example.com" {
		return data.ErrEditConflict
	}

	return nil

}
func (m *UserModel) GetByToken(scope, token string) (*data.User, error) {

	if scope != "activation" || token != ValidActivationToken {
		return nil, data.ErrRecordNotFound
	}

	return MockUser, nil
}
func (m *UserModel) GetByEmail(email string) (*data.User, error) {
	if email != "alice@example.com" {
		return nil, data.ErrRecordNotFound
	}

	return MockUser, nil
}
func (m *UserModel) GetByID(id int64) (*data.User, error) {

	if id != 1 {
		return nil, data.ErrRecordNotFound
	}

	return MockUser, nil

}
