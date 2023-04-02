package tea

import (
	"time"

	"code.gitea.io/sdk/gitea"
)

const ArchivedMirrorInterval = "0s"

type SourceRepository struct {
	Topics      []string
	Description string
	Private     bool
	Archived    bool
	PushedAt    time.Time
}

func (sr SourceRepository) StaleMirror(teaRepo *gitea.Repository) bool {
	return sr.PushedAt.After(teaRepo.MirrorUpdated)
}

func (sr SourceRepository) DiffDescription(teaRepo *gitea.Repository) bool {
	return teaRepo.Description != sr.Description
}

func (sr SourceRepository) DiffVisibility(teaRepo *gitea.Repository) bool {
	return teaRepo.Private != sr.Private
}

func (sr SourceRepository) DiffMirrorInterval(teaRepo *gitea.Repository) bool {
	if sr.Archived {
		return teaRepo.MirrorInterval != ArchivedMirrorInterval
	}

	return teaRepo.MirrorInterval == ArchivedMirrorInterval
}

func (sr SourceRepository) DiffTopics(teaTopics []string) bool {
Loop:
	for _, hubTopic := range sr.Topics {
		for _, teaTopic := range teaTopics {
			if teaTopic == hubTopic {
				continue Loop
			}
		}

		return true
	}

	return false
}
