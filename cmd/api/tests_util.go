package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	mockData "github.com/Kaungmyatkyaw2/book-store-api/internal/data/mock"
	mockMailer "github.com/Kaungmyatkyaw2/book-store-api/internal/mailer/mock"
	"github.com/hashicorp/go-hclog"
)

func newTestApplication(t *testing.T) *application {

	return &application{
		logger: hclog.Default(),
		models: data.Models{
			Users:  &mockData.UserModel{},
			Tokens: &mockData.TokenModel{},
		},
		mailer: mockMailer.Mailer{},
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, map[string]string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)

	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	var data map[string]string
	if err := json.NewDecoder(rs.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return rs.StatusCode, rs.Header, data
}

func (ts *testServer) post(t *testing.T, urlPath string, body interface{}) (int, http.Header, map[string]any) {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer rs.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(rs.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return rs.StatusCode, rs.Header, data
}

func structToMap(in interface{}) map[string]interface{} {
	b, _ := json.Marshal(in)

	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m
}
