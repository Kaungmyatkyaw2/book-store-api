package main

import (
	"net/http"
)

// GetUserByID godoc
// @Summary Get user by ID
// @Description Returns a single user
// @Tags users
// @Accept  json
// @Produce  json
// @Param   id  path  int  true  "User ID"
// @Success 200 {string} string "User found"
// @Failure 404 {string} string "User not found"
// @Router /api/users/{id} [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":      "available",
		"environment": app.config.env,
		// "version" : version,
	}

	err := app.writeJson(w, http.StatusOK, data, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
