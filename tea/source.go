package tea

import "strings"

type SourceRepository struct {
	SyncRepository
	Owner    string
	Name     string
	Fork     bool
	CloneURL string
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
