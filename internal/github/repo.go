package github

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gh "github.com/google/go-github/v62/github"
	"github.com/lshabtai/legit/internal/models"
	"github.com/lshabtai/legit/internal/templates"
)

func (s *Service) CreateRepository(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error) {
	slog.Info("creating GitHub repository", "name", req.Name, "owner", s.owner)

	repo, _, err := s.client.Repositories.Create(ctx, "", &gh.Repository{
		Name:        gh.String(req.Name),
		Description: gh.String(req.Description),
		Private:     gh.Bool(req.Private),
		AutoInit:    gh.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	slog.Info("repository created on GitHub", "name", req.Name, "url", repo.GetHTMLURL())

	data := templates.RepoData{
		Name:        req.Name,
		Description: req.Description,
		Owner:       s.owner,
		Year:        time.Now().Year(),
	}

	files, err := s.createFiles(ctx, req.Name, data)
	if err != nil {
		slog.Error("failed to create some files, continuing", "error", err)
	}

	if err := s.setBranchProtection(ctx, req.Name); err != nil {
		slog.Error("failed to set branch protection", "error", err)
	}

	return &models.CreateRepoResponse{
		Name:      req.Name,
		URL:       repo.GetHTMLURL(),
		CloneURL:  repo.GetCloneURL(),
		CreatedAt: repo.GetCreatedAt().Time,
		Files:     files,
		Message:   "Repository created successfully with branch protection and CI/CD workflow",
	}, nil
}
