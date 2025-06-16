package main

import (
	"net/http"
	"reflect"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
		wantBody   map[string]interface{}
	}{
		// {
		// 	name: "Valid Register request",
		// 	payload: map[string]string{
		// 		"name":     "John",
		// 		"email":    "john@example.com",
		// 		"password": "password123",
		// 	},
		// 	wantStatus: http.StatusAccepted,
		// 	wantBody: map[string]any{
		// 		"message": "user successfully registered",
		// 	},
		// },
		{
			name: "Missing email",
			payload: map[string]string{
				"name":     "Bob",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"error": map[string]any{
					"email": "must be provided",
				},
			},
		},
		{
			name: "Duplicate email",
			payload: map[string]string{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"error": map[string]any{
					"email": "a user with this email address is already registered",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, body := ts.post(t, "/v1/auth/register", tt.payload)

			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}

			if !reflect.DeepEqual(body, tt.wantBody) {
				t.Errorf("expected body = %#v, got %#v", tt.wantBody, body)
			}
		})
	}
}
