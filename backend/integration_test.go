package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBooksIntegration(t *testing.T) {
	app := setupFiberApp()

	req, err := http.NewRequest("GET", "/api/books", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	app.Handler(req, rr)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("error decoding response body: %v", err)
	}

	expectedMessage := "books fetched successfully"
	if message, ok := response["message"].(string); !ok || message != expectedMessage {
		t.Errorf("unexpected message: got %v, want %v", message, expectedMessage)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Errorf("unexpected data format in response")
	}

}

func setupFiberApp() *fiber.App {
	app := fiber.New()
	return app
}
