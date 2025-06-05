package main

// ErrorResponse represents the structure of all error messages
type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}
