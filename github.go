package main

import (
	"code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v50/github"
)

func isHubRepoMirror(teaRepo *gitea.Repository, hubRepo *github.Repository) bool {
	return teaRepo.Mirror && (hubRepo.GetCloneURL() == teaRepo.OriginalURL || hubRepo.GetHTMLURL() == teaRepo.OriginalURL)
}

func isHubRepoMirrorStale(teaRepo *gitea.Repository, hubRepo *github.Repository) bool {
	return hubRepo.PushedAt.After(teaRepo.MirrorUpdated)
}

func isHubRepoDescriptionDifferent(teaRepo *gitea.Repository, hubRepo *github.Repository) bool {
	return teaRepo.Description != hubRepo.GetDescription()
}

func isHubRepoVisibilityDifferent(teaRepo *gitea.Repository, hubRepo *github.Repository) bool {
	return teaRepo.Private != hubRepo.GetPrivate()
}

func isHubRepoMirrorIntervalDifferent(teaRepo *gitea.Repository, hubRepo *github.Repository) bool {
	if hubRepo.GetArchived() {
		return teaRepo.MirrorInterval != "0s"
	}
	return teaRepo.MirrorInterval == "0s"
}

func isHubTopicsDifferent(teaTopics []string, hubRepo *github.Repository) bool {
Loop:
	for _, hubTopic := range hubRepo.Topics {
		for _, teaTopic := range teaTopics {
			if teaTopic == hubTopic {
				continue Loop
			}
		}

		return true
	}

	return false
}
