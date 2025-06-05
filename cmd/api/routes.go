package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", (app.healthCheckHandler))

	router.HandlerFunc(http.MethodPost, "/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/auth/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/login", app.loginHandler)

	router.Handler("GET", "/swagger/*any", httpSwagger.WrapHandler)

	return router
}
