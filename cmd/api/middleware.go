package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {

			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidCredentialsResponse(w, r)
			return
		}

		token := headerParts[1]

		verifiedToken, err := app.verifyJWTToken(token)

		if err != nil {
			switch {
			case strings.HasPrefix(err.Error(), "invalid jwt"):
				app.invalidAuthenticationTokenResponse(w, r)
			case strings.Contains(err.Error(), "token is expired"):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)

			}
			return
		}

		var userID int64

		if claims, ok := verifiedToken.Claims.(jwt.MapClaims); ok {
			userID = int64(claims["userID"].(float64))
		}

		user, err := app.models.Users.GetByID(userID)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if user == nil {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}
