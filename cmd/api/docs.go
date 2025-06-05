package main

import (
	"time"
)

// Responses TYPE DTOS
type UserResponseDTO struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
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

// Responses
type InternalServerErrorResponse struct {
	Error string `json:"error"`
}

type GeneralErrorResponse struct {
	Error map[string]string `json:"error"`
}

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

type RegisterUserResponse struct {
	Data UserResponseDTO `json:"data"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
