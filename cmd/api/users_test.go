package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
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
		{
			name: "Valid request",
			payload: map[string]string{
				"name":     "John",
				"email":    "john@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusAccepted,
			wantBody: map[string]interface{}{
				"data": data.User{
					Name:  "John",
					Email: "john@example.com",
				},
			},
		},
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

			if tt.name == "Valid request" {

				wantData, _ := tt.wantBody["data"].(data.User)

				bodyData, ok := body["data"]

				jsonBytes, err := json.Marshal(bodyData)

				if err != nil {
					t.Fatal("Err while marshalling body", err.Error())
				}

				var bodyUserData data.User

				err = json.Unmarshal(jsonBytes, &bodyUserData)

				if err != nil {
					t.Fatal("Err while unmarshalling body", err.Error())
				}

				if !ok {
					t.Errorf("exepected data to be existed")
					return
				}

				if bodyUserData.Email != wantData.Email {
					t.Errorf("exepected email = %s, got %s", wantData.Email, bodyUserData.Email)
					return
				}
				if bodyUserData.Name != wantData.Name {
					t.Errorf("exepected name = %s, got %s", wantData.Name, bodyUserData.Name)
					return
				}
				return
			}

			if !reflect.DeepEqual(body, tt.wantBody) {
				t.Errorf("expected body = %#v, got %#v", tt.wantBody, body)
			}
		})
	}
}
