package main

import (
	"time"
)

// Responses TYPE DTOS

type MetadataDto struct {
	CurrentPage  int `json:"currentPage,omitempty"`
	PageSize     int `json:"pageSize,omitempty"`
	FirstPage    int `json:"firstPage,omitempty"`
	LastPage     int `json:"lastPage,omitempty"`
	TotalRecords int `json:"totalRecords,omitempty"`
}

type UserResponseDTO struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Activated    bool      `json:"activated"`
	AuthProvider string    `json:"authProvider"`
}

type BookResponseDTO struct {
	ID           int64      `json:"id"`
	Title        string     `json:"string"`
	UserID       int64      `json:"userId"`
	CoverPicture string     `json:"coverPicture"`
	CreatedAt    time.Time  `json:"createdAt"`
	IsPublished  bool       `json:"isPublished"`
	PublishedAt  *time.Time `json:"publishedAt"`
}

// Requests Parts

type RegisterUserRequestBody struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ActivateUserRequestBody struct {
	Token string `json:"token" binding:"required"`
}

type LoginRequestBody struct {
	Email    string `json:"email"  binding:"required"`
	Password string `json:"password"  binding:"required"`
}

type CreateBookBody struct {
	Title        string `json:"title"`
	CoverPicture string `json:"coverPicture"`
}

type UpdateBookBody struct {
	Title        string     `json:"title"`
	CoverPicture string     `json:"coverPicture"`
	IsPublished  bool       `json:"isPublished"`
	PublishedAt  *time.Time `json:"publishedAt"`
}

// Responses
type InternalServerErrorResponse struct {
	Error string `json:"error"`
}

type GeneralErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error map[string]string `json:"error"`
}

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

type RegisterUserResponse struct {
	Data UserResponseDTO `json:"data"`
}

type GoogleLoginResponse struct {
	Url string `json:"url"`
}

type LoginResponse struct {
	AccessToken string `json:"acessToken"`
}

type GetBooksResponse struct {
	Data     []BookResponseDTO `json:"data"`
	Metadata MetadataDto       `json:"metadata"`
}

type BookResponse struct {
	Data BookResponseDTO `json:"data"`
}

type DeleteSuccessResponse struct {
	Message string `json:"message"`
}
