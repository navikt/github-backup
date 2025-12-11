package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v80/github"
)

func ReposInOrgs(ctx context.Context, orgs []string, gitHubToken string) ([]*github.Repository, error) {
	client := github.NewClient(nil).WithAuthToken(gitHubToken)
	client.UserAgent = "navikt-github-backup/0.0.1 (+https://github.com/navikt/github-backup) " + client.UserAgent

	var repos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for _, org := range orgs {
		for {
			r, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
			if err != nil {
				return nil, fmt.Errorf("error fetching repos in org %q: %w", org, err)
			}
			repos = append(repos, r...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}

	return repos, nil
}
