package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
	"github.com/golang-jwt/jwt/v5"
)

// RegisterUser godoc
// @Summary Register account
// @Description Signup for an account
// @Tags Authentication
// @Param request body RegisterUserRequestBody true "User registration data"
// @Produce  json
// @Success 202 {object} RegisterUserResponse "User signed up success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/auth/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user.AuthProvider = data.CredentialAuthProvider

	err = app.models.Users.Insert(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address is already registered")
			app.failedValidationResponse(w, r, v.Errors)

		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {

		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}

	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"data": user}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// ActivateUser godoc
// @Summary Activate registered account
// @Description Activate for an account
// @Tags Authentication
// @Param request body ActivateUserRequestBody true "User activation data"
// @Produce  json
// @Success 200 {object} RegisterUserResponse "User activated success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/auth/activate [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByToken(data.ScopeActivation, input.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.models.Tokens.DeleteTokensByUser(data.ScopeActivation, user.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": user}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// LoginAccount godoc
// @Summary Log in to an account
// @Description Login to an account
// @Tags Authentication
// @Param request body LoginRequestBody true "Login data"
// @Produce  json
// @Success 200 {object} LoginResponse "Login success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
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
			app.badRequestResponse(w, r, errors.New("refresh token isn't found in cookies"))
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	verifiedToken, err := app.verifyJWTToken(rToken.Value)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "invalid jwt"):
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

	err = app.issueAccessToken(w, userID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
