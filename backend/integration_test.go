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
	// Set up the Fiber app with the same configuration as in your main function
	app := setupFiberApp()

	// Create HTTP request to the /api/books endpoint
	req, err := http.NewRequest("GET", "/api/books", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// Create ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Execute HTTP request using the app's handler and record the response
	app.Handler(req, rr)

	// Check the HTTP status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Parse the JSON response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("error decoding response body: %v", err)
	}

	// Check the response message
	expectedMessage := "books fetched successfully"
	if message, ok := response["message"].(string); !ok || message != expectedMessage {
		t.Errorf("unexpected message: got %v, want %v", message, expectedMessage)
	}

	// Check the data field (assuming it's an array)
	data, ok := response["data"].([]interface{})
	if !ok {
		t.Errorf("unexpected data format in response")
	}

	// Optionally check other fields in the response

	// ...

	// Optionally assert additional conditions based on your application's behavior

	// ...
}

func setupFiberApp() *fiber.App {
	// Initialize your Fiber app and add routes
	app := fiber.New()

	// Add your routes or set up your routes as in your main function

	return app
}
