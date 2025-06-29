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

	router.Handler("GET", "/docs/*any", httpSwagger.WrapHandler)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/auth/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/login", app.loginHandler)
	router.HandlerFunc(http.MethodGet, "/v1/auth/google", app.googleLoginHandler)
	router.HandlerFunc(http.MethodGet, "/v1/auth/google/callback", app.googleCallbackHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/refresh", app.refreshTokenHandler)
	router.HandlerFunc(http.MethodGet, "/v1/auth/me", app.requireActivatedUser(app.getMe))

	router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.getUser)
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/books", app.getBooksByUser)

	router.HandlerFunc(http.MethodGet, "/v1/books", app.getBooksHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/:id", app.getBookByIDHandler)
	router.HandlerFunc(http.MethodPost, "/v1/books", app.requireActivatedUser(app.createBookHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/books/:id", app.requireActivatedUser(app.updateBookHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/books/:id", app.requireActivatedUser(app.deleteBookHandler))
	router.HandlerFunc(http.MethodGet, "/v1/books/:id/chapters", app.getChaptersByBookHandler)

	router.HandlerFunc(http.MethodPost, "/v1/chapters", app.requireActivatedUser(app.createChapterHandler))

	return app.authenticate(router)
}
