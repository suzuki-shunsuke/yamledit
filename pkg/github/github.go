package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/go-github/v88/github"
	"github.com/suzuki-shunsuke/ghtkn-go-sdk/ghtkn"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"golang.org/x/oauth2"
)

type Client struct {
	repo RepositoriesService
}

type RepositoriesService interface {
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

type (
	RepositoryContentGetOptions = github.RepositoryContentGetOptions
)

func New(ctx context.Context, logger *slog.Logger, token string) (*Client, error) {
	gh, err := github.NewClient(github.WithHTTPClient(getHTTPClient(ctx, logger, token)))
	if err != nil {
		return nil, fmt.Errorf("create a GitHub client: %w", err)
	}
	return &Client{
		repo: gh.Repositories,
	}, nil
}

func getHTTPClient(ctx context.Context, logger *slog.Logger, token string) *http.Client {
	ts, err := getTokenSource(logger, token)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get a token source")
		return http.DefaultClient
	}
	if ts == nil {
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, ts)
}

func getTokenSource(logger *slog.Logger, token string) (oauth2.TokenSource, error) {
	if token != "" {
		return oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		), nil
	}
	ghtknEnabled, err := ghtkn.Enabled(&ghtkn.InputEnabled{
		Envs: []string{"YAMLEDIT_GHTKN_ENABLED"},
	})
	if err != nil {
		return nil, fmt.Errorf("get ghtkn enabled: %w", err)
	}
	if !ghtknEnabled {
		return nil, nil //nolint:nilnil
	}
	client, err := ghtkn.New()
	if err != nil {
		return nil, fmt.Errorf("create a ghtkn client: %w", err)
	}
	return client.TokenSource(logger, &ghtkn.InputGet{}), nil
}

func GetGitHubTokenFromEnv() string {
	for _, key := range []string{"YAMLEDIT_GITHUB_TOKEN", "GITHUB_TOKEN"} {
		s := os.Getenv(key)
		if s != "" {
			return s
		}
	}
	return ""
}
