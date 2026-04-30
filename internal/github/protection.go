package github

import (
	"context"
	"fmt"
	"log/slog"

	gh "github.com/google/go-github/v62/github"
)

func (s *Service) setBranchProtection(ctx context.Context, repoName string) error {
	slog.Debug("setting branch protection", "repo", repoName, "branch", "main")

	_, _, err := s.client.Repositories.UpdateBranchProtection(ctx, s.owner, repoName, "main", &gh.ProtectionRequest{
		RequiredPullRequestReviews: &gh.PullRequestReviewsEnforcementRequest{
			RequiredApprovingReviewCount: 1,
		},
		RequiredStatusChecks: &gh.RequiredStatusChecks{
			Strict:   true,
			Contexts: &[]string{"lint"},
		},
		EnforceAdmins: false,
	})
	if err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	return nil
}
