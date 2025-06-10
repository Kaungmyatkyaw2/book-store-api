package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

// RegisterUser godoc
// @Summary Register account
// @Description Signup for an account
// @Tags Authentication
// @Param request body RegisterUserRequestBody true "User registration data"
// @Produce  json
// @Success 202 {object} RegisterUserResponse "User signed up success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} GeneralErrorResponse "Validation Error"
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
// @Failure 422 {object} GeneralErrorResponse "Validation Error"
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
