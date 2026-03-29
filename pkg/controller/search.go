package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sahilm/fuzzy"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

type searchResult struct {
	Repositories []searchRepo `json:"repositories"`
}

type searchRepo struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
	License     string `json:"license"`
}

func Search(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, query string) error {
	repos, err := fetchTopicRepos(ctx, logger, ghClient, c)
	if err != nil {
		return err
	}
	if query != "" {
		repos = filterByQuery(repos, query)
	}
	for _, r := range repos {
		fmt.Fprintf(os.Stdout, "https://github.com/%s/%s\t%s\t%s\n", r.Owner, r.Name, r.License, r.Description)
	}
	return nil
}

func fetchTopicRepos(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache) ([]searchRepo, error) {
	if b, ok := c.GetTopicSearch(logger.Logger); ok {
		var result searchResult
		if err := json.Unmarshal(b, &result); err != nil {
			logger.Debug("parse cached topic search result", "error", err)
		} else {
			return result.Repositories, nil
		}
	}
	ghRepos, err := ghClient.SearchRepositories(ctx, "topic:yamledit-rule")
	if err != nil {
		return nil, fmt.Errorf("search repositories: %w", err)
	}
	repos := make([]searchRepo, 0, len(ghRepos))
	for _, r := range ghRepos {
		license := ""
		if r.License != nil && r.License.SPDXID != nil {
			license = *r.License.SPDXID
		}
		repos = append(repos, searchRepo{
			Owner:       r.GetOwner().GetLogin(),
			Name:        r.GetName(),
			Description: r.GetDescription(),
			License:     license,
		})
	}
	result := searchResult{Repositories: repos}
	b, err := json.Marshal(result)
	if err != nil {
		return repos, nil //nolint:nilerr
	}
	if err := c.PutTopicSearch(b); err != nil {
		logger.Debug("cache topic search result", "error", err)
	}
	return repos, nil
}

func filterByQuery(repos []searchRepo, query string) []searchRepo {
	strs := make([]string, len(repos))
	for i, r := range repos {
		strs[i] = r.Name + " " + r.Description
	}
	matches := fuzzy.Find(query, strs)
	filtered := make([]searchRepo, 0, len(matches))
	for _, m := range matches {
		filtered = append(filtered, repos[m.Index])
	}
	return filtered
}
