package github

import (
	"context"
	"fmt"
	"log/slog"

	gh "github.com/google/go-github/v62/github"
	"github.com/lshabtai/legit/internal/templates"
)

type fileEntry struct {
	path    string
	content string
}

func (s *Service) createFiles(ctx context.Context, repoName string, data templates.RepoData) ([]string, error) {
	files := []fileEntry{
		{path: "README.md", content: templates.Readme(data)},
		{path: ".gitignore", content: templates.Gitignore()},
		{path: "LICENSE", content: templates.License(data)},
		{path: "go.mod", content: templates.GoMod(data)},
		{path: "main.go", content: templates.Microservice(data)},
		{path: "Dockerfile", content: templates.ServiceDockerfile()},
		{path: ".github/workflows/ci.yml", content: templates.CIWorkflow()},
	}

	created := make([]string, 0, len(files))
	for _, f := range files {
		slog.Debug("creating file", "repo", repoName, "path", f.path)
		_, _, err := s.client.Repositories.CreateFile(ctx, s.owner, repoName, f.path, &gh.RepositoryContentFileOptions{
			Message: gh.String(fmt.Sprintf("chore: add %s", f.path)),
			Content: []byte(f.content),
		})
		if err != nil {
			return created, fmt.Errorf("failed to create %s: %w", f.path, err)
		}
		created = append(created, f.path)
	}

	return created, nil
}
