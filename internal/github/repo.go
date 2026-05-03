package github

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gh "github.com/google/go-github/v62/github"
	"github.com/sliron1989/repoistory-tool/internal/models"
	"github.com/sliron1989/repoistory-tool/internal/templates"
)

func (s *Service) CreateRepository(ctx context.Context, req *models.CreateRepoRequest) (*models.CreateRepoResponse, error) {
	slog.Info("creating GitHub repository", "name", req.Name, "owner", s.owner)

	// AutoInit deliberately omitted: GitHub's auto-init creates a starter
	// README.md, which then collides with our own README.md seed (the
	// "create file" API rejects writes to an existing path without a sha).
	// Letting our first CreateFile call be the initial commit avoids the
	// conflict and ensures our template content is what lands on main.
	repo, _, err := s.client.Repositories.Create(ctx, "", &gh.Repository{
		Name:        gh.String(req.Name),
		Description: gh.String(req.Description),
		Private:     gh.Bool(req.Private),
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
		slog.Error("file seeding failed", "name", req.Name, "url", repo.GetHTMLURL(), "created_files", files, "error", err)
		return nil, fmt.Errorf("repository %s was created at %s but file seeding failed after %d files (%v): %w",
			req.Name, repo.GetHTMLURL(), len(files), files, err)
	}

	if err := s.setBranchProtection(ctx, req.Name); err != nil {
		slog.Error("branch protection failed", "name", req.Name, "url", repo.GetHTMLURL(), "error", err)
		return nil, fmt.Errorf("repository %s was created at %s with all files but branch protection failed: %w",
			req.Name, repo.GetHTMLURL(), err)
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
