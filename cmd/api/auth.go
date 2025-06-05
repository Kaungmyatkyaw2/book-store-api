package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
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

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.createJWTToken(user.ID, time.Now().Add(time.Hour*24).Unix())

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"token": token}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
