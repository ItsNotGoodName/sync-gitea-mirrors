package tea

import (
	"code.gitea.io/sdk/gitea"
)

func ConvertList(teaRepos []*gitea.Repository, getTopic func(r *gitea.Repository) []string) []SourceRepository {
	repos := make([]SourceRepository, len(teaRepos))
	for i := range repos {
		repos[i] = Convert(teaRepos[i], getTopic(teaRepos[i]))
	}
	return repos
}

func Convert(r *gitea.Repository, topics []string) SourceRepository {
	return SourceRepository{
		SyncRepository: SyncRepository{
			Topics:      topics,
			Description: r.Description,
			Private:     r.Private,
			Archived:    r.Archived,
			PushedAt:    r.Updated,
		},
		Owner: r.Owner.UserName,
		Name:  r.Name,
		Fork:  r.Fork,
		URLS:  []string{r.CloneURL},
	}
}

func GetRepoOrNil(client *gitea.Client, owner, repoName string) (*gitea.Repository, error) {
	repo, teaRepoResp, err := client.GetRepo(owner, repoName)
	if err != nil {
		if teaRepoResp != nil && teaRepoResp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return repo, nil
}

func ListRepos(client *gitea.Client, owner string, skipPrivate bool, skipForks bool) ([]*gitea.Repository, error) {
	opts := gitea.ListOptions{Page: -1}
	var repos []*gitea.Repository
	if owner == "" {
		var err error
		repos, _, err = client.ListMyRepos(gitea.ListReposOptions{ListOptions: opts})
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		repos, _, err = client.ListUserRepos(owner, gitea.ListReposOptions{ListOptions: opts})
		if err != nil {
			return nil, err
		}
	}

	if !skipPrivate && !skipForks {
		return repos, nil
	}

	// Skip private or forks
	var newRepos []*gitea.Repository
	for _, repo := range repos {
		if skipPrivate && repo.Private {
			continue
		}

		if skipForks && repo.Fork {
			continue
		}

		newRepos = append(newRepos, repo)
	}

	return newRepos, nil
}
