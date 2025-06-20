package main

import (
	"errors"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
)

// GetMe godoc
// @Summary Get me
// @Description Get current logged in user information
// @Tags Authentication
// @Produce  json
// @Success 200 {object} GetUserResponse "Get Current Loggined User successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Router /v1/auth/me [get]
func (app *application) getMe(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	err := app.writeJSON(w, 200, envelope{"data": user}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// GetUserByID godoc
// @Summary Get user by id
// @Description Get specific user by id
// @Tags Users
// @Param id path int true "User ID"
// @Produce  json
// @Success 200 {object} GetUserResponse "Get Specific User successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 400 {object} GeneralErrorResponse "BadRequest Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/users/{id} [get]
func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.models.Users.GetByID(int64(id))

	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.badRequestResponse(w, r, errors.New("user id mismatched"))
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, 200, envelope{"data": user}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)

	}

}
