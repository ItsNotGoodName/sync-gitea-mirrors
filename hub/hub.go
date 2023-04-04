package hub

import (
	"context"
	"fmt"

	"github.com/ItsNotGoodName/sync-gitea-mirrors/tea"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

func ConvertList(hubRepos []*github.Repository) []tea.SourceRepository {
	repos := make([]tea.SourceRepository, len(hubRepos))
	for i := range repos {
		repos[i] = Convert(hubRepos[i])
	}
	return repos
}

func Convert(r *github.Repository) tea.SourceRepository {
	return tea.SourceRepository{
		SyncRepository: tea.SyncRepository{
			Topics:      r.Topics,
			Description: r.GetDescription(),
			Private:     r.GetPrivate(),
			Archived:    r.GetArchived(),
			PushedAt:    r.GetPushedAt().Time,
		},
		Owner: r.GetOwner().GetLogin(),
		Name:  r.GetName(),
		Fork:  r.GetFork(),
		URLS:  []string{r.GetCloneURL(), r.GetHTMLURL()},
	}
}

func ListRepos(ctx context.Context, client *github.Client, owner string, skipPrivate bool, skipForks bool) ([]*github.Repository, error) {
	visiblity := "all"
	if skipPrivate {
		visiblity = "public"
	}
	var repos []*github.Repository
	page := 1
	limit := 100
	for page != 0 {
		pagedRepos, resp, err := client.Repositories.List(ctx, owner,
			&github.RepositoryListOptions{
				Sort:        "created",
				Visibility:  visiblity,
				ListOptions: github.ListOptions{Page: page, PerPage: limit},
			})
		if err != nil {
			return nil, fmt.Errorf("could not list GitHub repos: %w", err)
		}
		repos = append(repos, pagedRepos...)
		page = resp.NextPage
	}

	if skipForks {
		var notForkRepos []*github.Repository
		for _, r := range repos {
			if !r.GetFork() {
				notForkRepos = append(notForkRepos, r)
			}
		}
		repos = notForkRepos
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
