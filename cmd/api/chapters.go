package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

func (app *application) getChaptersByBookHandler(w http.ResponseWriter, r *http.Request) {
	bookId, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	chapters, err := app.models.Chapters.GetByBookId(bookId)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": chapters}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createChapterHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	if user == nil {
		app.authenticationRequiredResponse(w, r)
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BookID      int64  `json:"bookId"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	chapter := &data.Chapter{
		Title:       input.Title,
		Description: input.Description,
		BookID:      input.BookID,
	}

	v := validator.New()

	if data.ValidateChapter(v, chapter); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	book, err := app.models.Books.Get(input.BookID)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.errorResponse(w, r, http.StatusBadRequest, "book not found")
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if book.UserID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	err = app.models.Chapters.Insert(chapter)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/chapters/%d", chapter.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": chapter}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
