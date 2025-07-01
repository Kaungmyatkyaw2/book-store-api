package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

// GetChaptersByBook godoc
// @Summary Get Specific Book's Chapters
// @Description Get Created Chapters By Specific Book
// @Tags Chapters
// @Produce  json
// @Param id path int true "Book ID"
// @Success 200 {object} GetChaptersResponse "Fetched Chapters successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 404 {object} GeneralErrorResponse "Content Not Found Error"
// @Router /v1/books/{id}/chapters [get]
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

// GetChapterById godoc
// @Summary Get Chapter By ID
// @Description Get Specific Chapter By ID
// @Tags Chapters
// @Produce  json
// @Param id path int true "Chapter ID"
// @Success 200 {object} GetChapterResponse "Fetched chapter success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 404 {object} GeneralErrorResponse "Book not found"
// @Router /v1/chapters/{id} [get]
func (app *application) getChapterByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	chapter, err := app.models.Chapters.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": chapter}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)

	}
}

// CreateChapter godoc
// @Summary Create Chapters
// @Description Create Chapters
// @Tags Chapters
// @Param request body CreateChapterBody true "Chapter data to create"
// @Produce  json
// @Success 200 {object} GetChapterResponse "Book creation success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 403 {object} GeneralErrorResponse "Permission Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/chapters [post]
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
		UserID:      user.ID,
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

// UpdateChapter godoc
// @Summary Update Chapter
// @Description Update Chapter
// @Tags Chapters
// @Param request body UpdateChapterBody true "Chapter data to update"
// @Param id path int true "Chapter ID"
// @Produce  json
// @Success 200 {object} BookResponse "Updated chapter successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 403 {object} GeneralErrorResponse "Permission Error"
// @Failure 422 {object} ValidationErrorResponse "Validation Error"
// @Router /v1/chapters/{id} [patch]
func (app *application) updateChapterHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	chapter, err := app.models.Chapters.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if chapter.UserID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Content     *string `json:"content"`
	}

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		chapter.Title = *input.Title
	}

	if input.Description != nil {
		chapter.Description = *input.Description
	}

	if input.Content != nil {
		chapter.Content = input.Content
	}

	v := validator.New()

	if data.ValidateChapter(v, chapter); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Chapters.Update(chapter)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": chapter}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Deletechapter godoc
// @Summary Delete Chapter
// @Description Delete Chapter
// @Tags Chapters
// @Produce  json
// @Param id path int true "Chapter ID"
// @Success 200 {object} DeleteSuccessResponse "Deleted chapter successfully"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 401 {object} GeneralErrorResponse "Unauthenticated Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Failure 403 {object} GeneralErrorResponse "Permission Error"
// @Router /v1/chapters/{id} [delete]
func (app *application) deleteChapterHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	chapter, err := app.models.Chapters.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if chapter.UserID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	err = app.models.Chapters.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "chapter successfully deleted"}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
