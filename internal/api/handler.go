package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/lshabtai/legit/internal/metrics"
	"github.com/lshabtai/legit/internal/models"
)

var repoNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}$`)

// RepoCreator is the interface that the handler uses to create repositories.
type RepoCreator interface {
	CreateRepository(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error)
	CheckConnectivity(ctx context.Context) error
}

type Handler struct {
	ghService RepoCreator
	startTime time.Time
	version   string

	mu      sync.Mutex
	history []models.RepoRecord
}

func NewHandler(ghService RepoCreator, version string) *Handler {
	return &Handler{
		ghService: ghService,
		startTime: time.Now(),
		version:   version,
		history:   make([]models.RepoRecord, 0),
	}
}

// CreateRepo creates a new GitHub repository with standard files, CI/CD, and branch protection.
// @Summary Create a new GitHub repository
// @Description Creates a GitHub repository with README, .gitignore, LICENSE, Go microservice, Dockerfile, CI workflow, and branch protection rules
// @Tags repositories
// @Accept json
// @Produce json
// @Param request body models.CreateRepoRequest true "Repository creation request"
// @Success 201 {object} models.CreateRepoResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/repos [post]
func (h *Handler) CreateRepo(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body",
			Code:  http.StatusBadRequest,
		})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "name is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	if !repoNameRegex.MatchString(req.Name) {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid repository name: must start with alphanumeric and contain only alphanumeric, hyphens, dots, or underscores (max 100 chars)",
			Code:  http.StatusBadRequest,
		})
		return
	}

	if len(req.Description) > 350 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "description must be 350 characters or less",
			Code:  http.StatusBadRequest,
		})
		return
	}

	metrics.ActiveRepoCreations.Inc()
	start := time.Now()

	slog.Info("creating repository", "name", req.Name, "private", req.Private)

	resp, err := h.ghService.CreateRepository(r.Context(), &req)
	duration := time.Since(start)

	metrics.ActiveRepoCreations.Dec()

	if err != nil {
		slog.Error("failed to create repository", "name", req.Name, "error", err, "duration_ms", duration.Milliseconds())
		metrics.ReposCreated.WithLabelValues("error").Inc()
		metrics.RepoCreationDuration.WithLabelValues("error").Observe(duration.Seconds())
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	metrics.ReposCreated.WithLabelValues("success").Inc()
	metrics.RepoCreationDuration.WithLabelValues("success").Observe(duration.Seconds())

	h.mu.Lock()
	h.history = append(h.history, models.RepoRecord{
		Name:        resp.Name,
		URL:         resp.URL,
		CloneURL:    resp.CloneURL,
		Description: req.Description,
		Private:     req.Private,
		CreatedAt:   resp.CreatedAt,
	})
	h.mu.Unlock()

	slog.Info("repository created", "name", resp.Name, "url", resp.URL, "duration_ms", duration.Milliseconds())

	writeJSON(w, http.StatusCreated, resp)
}

// ListRepos returns all repositories created in the current session.
// @Summary List created repositories
// @Description Returns all repositories created since the service started
// @Tags repositories
// @Produce json
// @Success 200 {array} models.RepoRecord
// @Router /api/v1/repos [get]
func (h *Handler) ListRepos(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	repos := make([]models.RepoRecord, len(h.history))
	copy(repos, h.history)
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, repos)
}

// Health returns the service health status.
// @Summary Health check
// @Description Returns service status, version, and uptime
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.HealthResponse{
		Status:  "ok",
		Version: h.version,
		Uptime:  time.Since(h.startTime).Round(time.Second).String(),
	})
}

// Ready checks GitHub API connectivity.
// @Summary Readiness check
// @Description Verifies the service can connect to the GitHub API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} models.ErrorResponse
// @Router /readyz [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.ghService.CheckConnectivity(r.Context()); err != nil {
		slog.Warn("readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "github connectivity check failed: " + err.Error(),
			Code:  http.StatusServiceUnavailable,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
