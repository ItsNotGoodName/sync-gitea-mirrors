package tea

import (
	"code.gitea.io/sdk/gitea"
)

func GetRepo(client *gitea.Client, owner, repoName string) (*gitea.Repository, error) {
	repo, teaRepoResp, err := client.GetRepo(owner, repoName)
	if err != nil {
		if teaRepoResp != nil && teaRepoResp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return repo, nil
}
