package mock

import (
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
)

var (
	MockTokenPlaintext = "ABCDEFGHIJKLMNOPQRSTUVWX"
	mockTokenHash      = []byte("fakehashedtokenforuser")
)

type TokenModel struct {
	Tokens []data.Token
}

func (m *TokenModel) New(userID int64, ttl time.Duration, scope string) (*data.Token, error) {
	token := &data.Token{
		Plaintext: MockTokenPlaintext,
		Hash:      mockTokenHash,
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	m.Tokens = append(m.Tokens, *token)
	return token, nil
}

func (m *TokenModel) Insert(token *data.Token) error {
	m.Tokens = append(m.Tokens, *token)
	return nil
}

func (m *TokenModel) DeleteTokensByUser(scope string, userID int64) error {

	filtered := []data.Token{}
	for _, t := range m.Tokens {
		if !(t.Scope == scope && t.UserID == userID) {
			filtered = append(filtered, t)
		}
	}
	m.Tokens = filtered
	return nil
}
