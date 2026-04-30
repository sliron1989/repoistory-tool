package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lshabtai/legit/internal/models"
)

// mockRepoCreator implements RepoCreator for testing.
type mockRepoCreator struct {
	createFunc      func(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error)
	connectivityErr error
}

func (m *mockRepoCreator) CreateRepository(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return &models.CreateRepoResponse{
		Name:      req.Name,
		URL:       "https://github.com/testuser/" + req.Name,
		CloneURL:  "https://github.com/testuser/" + req.Name + ".git",
		CreatedAt: time.Now(),
		Files:     []string{"README.md", ".gitignore", "LICENSE", "go.mod", "main.go", "Dockerfile", ".github/workflows/ci.yml"},
		Message:   "Repository created successfully",
	}, nil
}

func (m *mockRepoCreator) CheckConnectivity(ctx context.Context) error {
	return m.connectivityErr
}

func newTestHandler() *Handler {
	return NewHandler(&mockRepoCreator{}, "test")
}

func TestHealth(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp models.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Version != "test" {
		t.Errorf("expected version 'test', got %q", resp.Version)
	}
	if resp.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestReady_Success(t *testing.T) {
	h := NewHandler(&mockRepoCreator{connectivityErr: nil}, "test")
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	h.Ready(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestReady_Failure(t *testing.T) {
	h := NewHandler(&mockRepoCreator{connectivityErr: fmt.Errorf("connection refused")}, "test")
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	h.Ready(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}
}

func TestCreateRepo_Success(t *testing.T) {
	h := newTestHandler()
	body := `{"name":"my-service","description":"A test service","private":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateRepo(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.CreateRepoResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "my-service" {
		t.Errorf("expected name 'my-service', got %q", resp.Name)
	}
	if len(resp.Files) != 7 {
		t.Errorf("expected 7 files, got %d", len(resp.Files))
	}
}

func TestCreateRepo_EmptyName(t *testing.T) {
	h := newTestHandler()
	body := `{"name":"","description":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateRepo(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateRepo_InvalidName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"-starts-with-dash"},
		{".starts-with-dot"},
		{"has spaces"},
		{"has/slash"},
		{"has@symbol"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			body := fmt.Sprintf(`{"name":%q}`, tt.name)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
			w := httptest.NewRecorder()

			h.CreateRepo(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("name %q: expected 400, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestCreateRepo_ValidNames(t *testing.T) {
	tests := []string{
		"my-service",
		"my_service",
		"my.service",
		"MyService123",
		"a",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			h := newTestHandler()
			body := fmt.Sprintf(`{"name":%q}`, name)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
			w := httptest.NewRecorder()

			h.CreateRepo(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("name %q: expected 201, got %d; body: %s", name, w.Code, w.Body.String())
			}
		})
	}
}

func TestCreateRepo_DescriptionTooLong(t *testing.T) {
	h := newTestHandler()
	longDesc := strings.Repeat("a", 351)
	body := fmt.Sprintf(`{"name":"test","description":%q}`, longDesc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateRepo(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateRepo_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.CreateRepo(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateRepo_GitHubError(t *testing.T) {
	mock := &mockRepoCreator{
		createFunc: func(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error) {
			return nil, fmt.Errorf("GitHub API error: 422 Validation Failed")
		},
	}
	h := NewHandler(mock, "test")
	body := `{"name":"test-repo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateRepo(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "GitHub API error") {
		t.Errorf("expected error to contain 'GitHub API error', got %q", resp.Error)
	}
}

func TestListRepos_Empty(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil)
	w := httptest.NewRecorder()

	h.ListRepos(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var repos []models.RepoRecord
	json.NewDecoder(w.Body).Decode(&repos)
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestListRepos_AfterCreate(t *testing.T) {
	h := newTestHandler()

	// Create a repo first
	body := `{"name":"test-repo","description":"test"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewBufferString(body))
	createW := httptest.NewRecorder()
	h.CreateRepo(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", createW.Code)
	}

	// Now list
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil)
	listW := httptest.NewRecorder()
	h.ListRepos(listW, listReq)

	var repos []models.RepoRecord
	json.NewDecoder(listW.Body).Decode(&repos)
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "test-repo" {
		t.Errorf("expected name 'test-repo', got %q", repos[0].Name)
	}
}

func TestWriteJSON_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}
