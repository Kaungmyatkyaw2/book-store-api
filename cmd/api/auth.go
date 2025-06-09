package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
	"github.com/golang-jwt/jwt/v5"
)

// LoginAccount godoc
// @Summary Log in to an account
// @Description Login to an account
// @Tags Authentication
// @Param request body LoginRequestBody true "Login data"
// @Produce  json
// @Success 200 {object} LoginResponse "Login success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} GeneralErrorResponse "Validation Error"
// @Failure 401 {object} GeneralErrorResponse "Invalid Credential Error"
// @Router /v1/auth/login [post]
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return

	}

	if user.AuthProvider != data.CredentialAuthProvider {
		app.badRequestResponse(w, r, errors.New("your account is not registerd with credentials"))
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	err = app.issueAccessToken(w, user.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// GoogleLogin godoc
// @Summary Refresh access token
// @Description Refresh Previous Access Token
// @Tags Authentication
// @Produce  json
// @Success 200 {object} LoginResponse "Return Redirect URL to continue Login with google"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Router /v1/auth/refresh [post]
func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	rToken, err := r.Cookie("jwt")

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.badRequestResponse(w, r, errors.New("refresh token isn't found in cookies."))
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	verifiedToken, err := app.verifyJWTToken(rToken.Value)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "invalid jwt"):
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	var userID int64

	if claims, ok := verifiedToken.Claims.(jwt.MapClaims); ok {
		userID = int64(claims["userID"].(float64))
	}

	err = app.issueAccessToken(w, userID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
