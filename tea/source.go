package tea

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
