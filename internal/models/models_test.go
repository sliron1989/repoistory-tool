package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateRepoRequest_JSON(t *testing.T) {
	input := `{"name":"my-svc","description":"A service","private":true}`
	var req CreateRepoRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Name != "my-svc" {
		t.Errorf("expected name 'my-svc', got %q", req.Name)
	}
	if req.Description != "A service" {
		t.Errorf("expected description 'A service', got %q", req.Description)
	}
	if !req.Private {
		t.Error("expected private=true")
	}
}

func TestCreateRepoResponse_JSON(t *testing.T) {
	resp := CreateRepoResponse{
		Name:      "test",
		URL:       "https://github.com/user/test",
		CloneURL:  "https://github.com/user/test.git",
		CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Files:     []string{"README.md", "main.go"},
		Message:   "done",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CreateRepoResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "test" {
		t.Errorf("expected name 'test', got %q", decoded.Name)
	}
	if len(decoded.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(decoded.Files))
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{Error: "not found", Code: 404}
	data, _ := json.Marshal(resp)

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Error != "not found" {
		t.Errorf("expected 'not found', got %q", decoded.Error)
	}
	if decoded.Code != 404 {
		t.Errorf("expected 404, got %d", decoded.Code)
	}
}

func TestHealthResponse_JSON(t *testing.T) {
	resp := HealthResponse{Status: "ok", Version: "1.0.0", Uptime: "5s"}
	data, _ := json.Marshal(resp)

	var decoded HealthResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != "ok" {
		t.Errorf("expected 'ok', got %q", decoded.Status)
	}
}
