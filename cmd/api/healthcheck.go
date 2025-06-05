package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
