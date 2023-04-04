package tea

import (
	"errors"
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
)

const ArchivedMirrorInterval = "0s"

type SyncRepository struct {
	Topics      []string
	Description string
	Private     bool
	Archived    bool
	PushedAt    time.Time
}

func (sr SyncRepository) StaleMirror(teaRepo *gitea.Repository) bool {
	return sr.PushedAt.After(teaRepo.MirrorUpdated)
}

func (sr SyncRepository) DiffDescription(teaRepo *gitea.Repository) bool {
	return teaRepo.Description != sr.Description
}

func (sr SyncRepository) DiffVisibility(teaRepo *gitea.Repository) bool {
	return teaRepo.Private != sr.Private
}

func (sr SyncRepository) DiffMirrorInterval(teaRepo *gitea.Repository) bool {
	if sr.Archived {
		return teaRepo.MirrorInterval != ArchivedMirrorInterval
	}

	return teaRepo.MirrorInterval == ArchivedMirrorInterval
}

func (sr SyncRepository) DiffTopics(teaTopics []string) bool {
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

type SyncConfig struct {
	SyncDescription    bool
	SyncVisibility     bool
	SyncTopics         bool
	SyncMirrorInterval bool
	DestMirrorInterval string
}

type SyncOutput struct {
	UpdateDescription    bool
	UpdateTopics         bool
	UpdateVisibility     bool
	UpdateMirrorInterval bool
	SyncMirror           bool
}

func Sync(client *gitea.Client, teaRepo *gitea.Repository, sourceRepo *SyncRepository, config SyncConfig) (SyncOutput, error) {
	owner := teaRepo.Owner.UserName
	repoName := teaRepo.Name

	var output SyncOutput
	var reterr error

	// Sync Description, MirrorInterval, Visibility
	{
		var archivedMirrorInterval = ArchivedMirrorInterval
		editRepoOption := gitea.EditRepoOption{}
		shouldEditRepo := false

		if config.SyncDescription && sourceRepo.DiffDescription(teaRepo) {
			editRepoOption.Description = &sourceRepo.Description

			output.UpdateDescription = true
			shouldEditRepo = true
		}

		if config.SyncVisibility && sourceRepo.DiffVisibility(teaRepo) {
			editRepoOption.Private = &sourceRepo.Private

			output.UpdateVisibility = true
			shouldEditRepo = true
		}

		if config.SyncMirrorInterval && sourceRepo.DiffMirrorInterval(teaRepo) {
			if sourceRepo.Archived {
				editRepoOption.MirrorInterval = &archivedMirrorInterval
			} else {
				editRepoOption.MirrorInterval = &config.DestMirrorInterval
			}

			output.UpdateMirrorInterval = true
			shouldEditRepo = true
		}

		if shouldEditRepo {
			_, _, err := client.EditRepo(owner, repoName, editRepoOption)
			if err != nil {
				reterr = errors.Join(reterr, fmt.Errorf("could not edit repo: %w", err))
				output.UpdateDescription = false
				output.UpdateVisibility = false
				output.UpdateMirrorInterval = false
			}
		}
	}

	// Sync Topics
	if config.SyncTopics {
		if teaTopics, _, err := client.ListRepoTopics(owner, repoName, gitea.ListRepoTopicsOptions{}); err != nil {
			reterr = errors.Join(reterr, fmt.Errorf("could not get repo topics: %w", err))
		} else if sourceRepo.DiffTopics(teaTopics) {
			if _, err := client.SetRepoTopics(owner, repoName, sourceRepo.Topics); err != nil {
				reterr = errors.Join(reterr, fmt.Errorf("could not set repo topics: %w", err))
			} else {
				output.UpdateTopics = true
			}
		}
	}

	// Handle cases where the source had commits after it was archived
	if sourceRepo.StaleMirror(teaRepo) {
		_, err := client.MirrorSync(owner, repoName)
		if err != nil {
			reterr = errors.Join(reterr, fmt.Errorf("could not mirror sync: %w", err))
		} else {
			output.SyncMirror = true
		}
	}

	return output, reterr
}
