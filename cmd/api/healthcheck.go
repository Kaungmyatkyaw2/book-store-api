package main

import (
	"net/http"
)

// HealthCheck godoc
// @Summary Health Check The API
// @Description Returns an object that include environment and status of the API
// @Tags Healthcheck
// @Produce  json
// @Success 200 {object} HealthCheckResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /v1/healthcheck [get]
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
