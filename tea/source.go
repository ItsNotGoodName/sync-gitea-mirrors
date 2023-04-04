package tea

import (
	"strings"

	"code.gitea.io/sdk/gitea"
)

type SourceRepository struct {
	SyncRepository
	Owner string
	Name  string
	Fork  bool
	URLS  []string
}

func (sr SourceRepository) GetFullName() string {
	return sr.Owner + "/" + sr.Name
}

func (sr SourceRepository) Is(nameOrPath string) bool {
	nameOrPath = strings.ToLower(nameOrPath)
	repoName := strings.ToLower(sr.Name)
	repoFullName := strings.ToLower(sr.GetFullName())
	if (nameOrPath == repoFullName) || (nameOrPath == repoName) {
		return true
	}

	_, after, _ := strings.Cut(nameOrPath, "/")
	return after == repoName
}

func (sr SourceRepository) IsMyMirror(teaRepo *gitea.Repository) bool {
	if !teaRepo.Mirror {
		return false
	}

	for _, url := range sr.URLS {
		if teaRepo.OriginalURL == url {
			return true
		}
	}

	return false

}
