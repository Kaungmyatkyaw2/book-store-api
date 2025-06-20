package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

// GetAllBooks godoc
// @Summary Get All Books
// @Description Get All Created Books
// @Tags Books
// @Produce  json
// @Param        page   query     int     false  "Page number (default: 1)"
// @Param        limit  query     int     false  "Items per page (default: 10)"
// @Param        sort   query     string  false  "Sort by field, e.g. 'name' or '-createdAt' for descending"
// @Success 200 {object} GetBooksResponse "Fetched Books successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Router /v1/books [get]
func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "limit", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "created_at")

	input.Filters.SortSafelist = []string{"id", "title", "created_at", "published_at", "-id", "-title", "-created_at", "-published_at"}

	if data.ValidateFilter(v, input.Filters); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := app.models.Books.GetAll(input.Title, input.Filters)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": books, "metadata": metadata}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// GetBooksByUser godoc
// @Summary Get All Books By User
// @Description Get Created Books By Specific User
// @Tags Users
// @Produce  json
// @Param id path int true "Book ID"
// @Param        page   query     int     false  "Page number (default: 1)"
// @Param        limit  query     int     false  "Items per page (default: 10)"
// @Param        sort   query     string  false  "Sort by field, e.g. 'name' or '-createdAt' for descending"
// @Success 200 {object} GetBooksResponse "Fetched Books successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 404 {object} GeneralErrorResponse "Content Not Found Error"
// @Router /v1/users/{id}/books [get]
func (app *application) getBooksByUser(w http.ResponseWriter, r *http.Request) {

	userID, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Title string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "limit", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "created_at")

	input.Filters.SortSafelist = []string{"id", "title", "created_at", "published_at", "-id", "-title", "-created_at", "-published_at"}

	if data.ValidateFilter(v, input.Filters); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := app.models.Books.GetAllByUser(input.Title, input.Filters, userID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": books, "metadata": metadata}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// GetBookById godoc
// @Summary Get Book By ID
// @Description Get Specific Book By ID
// @Tags Books
// @Produce  json
// @Param id path int true "Book ID"
// @Success 200 {object} BookResponse "User activated success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 404 {object} GeneralErrorResponse "Book not found"
// @Router /v1/books/{id} [get]
func (app *application) getBookByIDHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": book}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// CreateBook godoc
// @Summary Create Books
// @Description Create Books
// @Tags Books
// @Param request body CreateBookBody true "Book data to create"
// @Produce  json
// @Success 200 {object} BookResponse "Book creation success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/books [post]
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

// UpdateBook godoc
// @Summary Update Book
// @Description Update Book
// @Tags Books
// @Param request body UpdateBookBody true "Book data to update"
// @Param id path int true "Book ID"
// @Produce  json
// @Success 200 {object} BookResponse "Updated book successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 403 {object} GeneralErrorResponse "Permission Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/books/{id} [patch]
func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if book.UserID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Title        *string `json:"title"`
		CoverPicture *string `json:"coverPicture"`
		IsPublished  *bool   `json:"isPublished"`
	}

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateBook(v, book); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Update(book)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": book}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Deletebook godoc
// @Summary Delete Book
// @Description Delete Book
// @Tags Books
// @Produce  json
// @Param id path int true "Book ID"
// @Success 200 {object} DeleteSuccessResponse "Deleted book successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 403 {object} GeneralErrorResponse "Permission Error"
// @Router /v1/books/{id} [delete]
func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if book.UserID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	err = app.models.Books.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "book successfully deleted"}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
