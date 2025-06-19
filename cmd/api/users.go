package main

import "net/http"

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
