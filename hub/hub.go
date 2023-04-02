package hub

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

func GetRepos(ctx context.Context, client *github.Client, owner string) ([]*github.Repository, error) {
	var repos []*github.Repository
	page := 1
	limit := 100
	for page != 0 {
		pagedRepos, resp, err := client.Repositories.List(ctx, owner,
			&github.RepositoryListOptions{
				Sort:        "created",
				ListOptions: github.ListOptions{Page: page, PerPage: limit},
			})
		if err != nil {
			return nil, fmt.Errorf("could not list GitHub repos: %w", err)
		}
		repos = append(repos, pagedRepos...)
		page = resp.NextPage
	}

	return repos, nil
}

func NewClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func Migrate(client *gitea.Client, opt gitea.MigrateRepoOption, hubRepo *github.Repository, token string) error {
	opt.Service = gitea.GitServiceGithub
	opt.CloneAddr = hubRepo.GetCloneURL()
	opt.AuthToken = token
	_, _, err := client.MigrateRepo(opt)
	return err
}
