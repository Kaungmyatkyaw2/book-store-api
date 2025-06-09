package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) createJWTToken(userID, expDate int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    expDate,
	})

	tokenString, err := token.SignedString([]byte(app.config.jwt.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (app *application) setRefreshTokenCookie(w http.ResponseWriter, userID int64) error {
	refreshToken, err := app.createJWTToken(userID, time.Now().Add(time.Hour*24*7).Unix())

	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

	return nil
}

func (app *application) verifyJWTToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.config.jwt.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (app *application) issueAccessToken(w http.ResponseWriter, userID int64) error {
	accessToken, err := app.createJWTToken(userID, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		return err
	}

	if err := app.setRefreshTokenCookie(w, userID); err != nil {
		return err
	}

	return app.writeJson(w, http.StatusOK, envelope{"accessToken": accessToken}, nil)
}
