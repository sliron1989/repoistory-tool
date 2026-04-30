package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

type Service struct {
	client *gh.Client
	owner  string
}

func NewService(token string) (*Service, error) {
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := gh.NewClient(tc)

	// Get authenticated user to determine the owner
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	return &Service{
		client: client,
		owner:  user.GetLogin(),
	}, nil
}

func (s *Service) CheckConnectivity(ctx context.Context) error {
	_, _, err := s.client.Users.Get(ctx, "")
	return err
}
