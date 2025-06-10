package main

import (
	"fmt"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {

	books, err := app.models.Books.GetAll()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": books}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	if user == nil {
		app.authenticationRequiredResponse(w, r)
		return
	}

	var input struct {
		Title        string `json:"title"`
		CoverPicture string `json:"coverPicture"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.logger.Info("UseID: ", user.ID)

	book := &data.Book{
		Title:        input.Title,
		CoverPicture: input.CoverPicture,
		UserID:       user.ID,
	}

	v := validator.New()

	if data.ValidateBook(v, book); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Insert(book)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/books/%d", book.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": book}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
