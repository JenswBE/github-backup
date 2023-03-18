package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
)

func ListRepos(ctx context.Context, personalAccessToken string) ([]*github.Repository, error) {
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: personalAccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List repo's
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list GitHub repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

func ExtractRepoNames(repos []*github.Repository) []string {
	return lo.FilterMap(repos, func(r *github.Repository, _ int) (string, bool) { return r.GetName(), r.GetName() != "" })
}
