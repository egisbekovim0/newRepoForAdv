package main

import "testing"

func TestGenerateJWT(t *testing.T) {
  user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Role:  "user",
  }
  tokenString, err := GenerateJWT(user)
  if err != nil {
    t.Fatalf("Expected no error, got %v", err)
  }
  if tokenString == "" {
    t.Fatalf("Expected token to be generated, got empty string")
  }
  // Further tests could decode the token and check if the claims are correct.
}