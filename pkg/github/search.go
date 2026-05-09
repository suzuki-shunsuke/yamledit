package github

import (
	"context"
	"fmt"
)

func (c *Client) SearchRepositories(ctx context.Context, query string) ([]*Repository, error) {
	result, _, err := c.search.Repositories(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("search repositories by GitHub API: %w", err)
	}
	return result.Repositories, nil
}
