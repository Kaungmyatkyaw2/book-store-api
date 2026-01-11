package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/docs/*", httpSwagger.WrapHandler)

	r.Get("/v1/healthcheck", app.healthCheckHandler)

	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", app.registerUserHandler)
		r.Put("/activate", app.activateUserHandler)
		r.Post("/login", app.loginHandler)
		r.Get("/google", app.googleLoginHandler)
		r.Get("/google/callback", app.googleCallbackHandler)
		r.Post("/refresh", app.refreshTokenHandler)
		r.With(app.requireActivatedUser).Get("/me", app.getMe)
	})

	r.Route("/v1/users", func(r chi.Router) {
		r.Get("/{id}", app.getUser)
		r.With(app.requireActivatedUser).Get("/me/books", app.getBooksByCurrentUser)
		r.Get("/{id}/books", app.getBooksByUser)
	})

	r.Route("/v1/books", func(r chi.Router) {
		r.Get("/", app.getBooksHandler)
		r.With(app.requireActivatedUser).Get("/{id}", app.getBookByIDHandler)
		r.With(app.requireActivatedUser).Post("/", app.createBookHandler)
		r.With(app.requireActivatedUser).Patch("/{id}", app.updateBookHandler)
		r.With(app.requireActivatedUser).Delete("/{id}", app.deleteBookHandler)
		r.Get("/{id}/chapters", app.getChaptersByBookHandler)
	})

	r.Route("/v1/chapters", func(r chi.Router) {
		r.With(app.requireActivatedUser).Post("/", app.createChapterHandler)
		r.Get("/{id}", app.getChapterByIDHandler)
		r.With(app.requireActivatedUser).Patch("/{id}", app.updateChapterHandler)
		r.With(app.requireActivatedUser).Delete("/{id}", app.deleteChapterHandler)
	})

	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.methodNotAllowedResponse)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(r))))
}
