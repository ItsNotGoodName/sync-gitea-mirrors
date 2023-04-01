package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ItsNotGoodName/sync-gitea-mirrors/config"
	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"code.gitea.io/sdk/gitea"
)

var sugar *zap.SugaredLogger

func main() {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()

	// Parse config
	cfg := config.Config{}
	if err := config.Parse(&cfg); err != nil {
		sugar.Fatalln("could not parse config:", err)
	}

	// Create clients
	srcClient := NewGitHubClient(ctx, cfg.SrcToken)
	destClient, err := gitea.NewClient(cfg.DestURL, gitea.SetToken(cfg.DestToken))
	if err != nil {
		sugar.Fatalln("could not create Gitea client:", err)
	}

	// List GitHub repos
	var repos []*github.Repository
	{
		page := 1
		limit := 100
		for page != 0 {
			pagedRepos, resp, err := srcClient.Repositories.List(ctx, cfg.SrcOwner, &github.RepositoryListOptions{Sort: "created", ListOptions: github.ListOptions{Page: page, PerPage: limit}})
			if err != nil {
				sugar.Fatalln("could not list GitHub repos:", err)
			}
			repos = append(repos, pagedRepos...)
			page = resp.NextPage
		}
	}

	syncConfig := SyncConfig{
		SrcToken:           cfg.SrcToken,
		SyncDescription:    cfg.SyncDescription,
		SyncMirrorInterval: cfg.SyncMirrorInterval,
		SyncTopics:         cfg.SyncTopics,
		SyncVisibility:     cfg.SyncVisibility,
		DestOwner:          cfg.DestOwner,
		DestMirrorInterval: cfg.DestMirrorInterval,
	}

	for _, r := range repos {
		fmt.Printf("Syncing %s\n", r.GetFullName())
		if err := sync(&syncConfig, destClient, r); err != nil {
			sugar.Errorf("could not sync %s: %s\n", r.GetFullName(), err)
		}
	}
}

func NewGitHubClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

type SyncConfig struct {
	SrcToken           string
	SyncTopics         bool
	SyncDescription    bool
	SyncVisibility     bool
	SyncMirrorInterval bool
	DestOwner          string
	DestMirrorInterval string
}

func sync(cfg *SyncConfig, client *gitea.Client, hubRepo *github.Repository) error {
	// Decide on repo owner and repo name for Gitea
	owner := cfg.DestOwner
	if owner == "" {
		owner = hubRepo.Owner.GetLogin()
	}
	repoName := hubRepo.GetName()

	// Find Gitea repo or create new migration
	var teaRepo *gitea.Repository
	{
		var teaRepoResp *gitea.Response
		var err error
		teaRepo, teaRepoResp, err = client.GetRepo(owner, repoName)
		if err != nil {
			if teaRepoResp.StatusCode != 404 {
				return err
			}

			sugar.Info("new migration")

			_, _, err := client.MigrateRepo(gitea.MigrateRepoOption{
				Service:   gitea.GitServiceGithub,
				CloneAddr: hubRepo.GetCloneURL(),
				AuthToken: cfg.SrcToken,
				Mirror:    true,
				RepoOwner: owner,
				RepoName:  repoName,
			})

			return err
		}
	}

	var reterr error

	// Make sure Gitea repo is the mirror of the GitHub repo
	if !isHubRepoMirror(teaRepo, hubRepo) {
		return fmt.Errorf("not a mirror: %s", teaRepo.HTMLURL)
	}

	{
		editRepoOption := gitea.EditRepoOption{}
		shouldEditRepo := false

		if cfg.SyncDescription {
			sugar.Debug("description sync enabled")
			if isHubRepoDescriptionDifferent(teaRepo, hubRepo) {
				sugar.Infow("queue updating description update", "description", hubRepo.Description)
				editRepoOption.Description = hubRepo.Description
				shouldEditRepo = true
			}
		} else {
			sugar.Debug("description sync disabled")
		}

		if cfg.SyncVisibility {
			sugar.Debug("visibility sync enabled")
			if isHubRepoVisibilityDifferent(teaRepo, hubRepo) {
				sugar.Infow("queue visibility update", "private", hubRepo.Private)
				editRepoOption.Private = hubRepo.Private
				shouldEditRepo = true
			}
		} else {
			sugar.Debug("visibility sync disabled")
		}

		if cfg.SyncMirrorInterval {
			sugar.Debug("mirror interval sync enabled")
			if isHubRepoMirrorIntervalDifferent(teaRepo, hubRepo) {
				sugar.Infow("queue mirror interval update", "mirror_interval", cfg.DestMirrorInterval)
				editRepoOption.MirrorInterval = &cfg.DestMirrorInterval
				shouldEditRepo = true
			}
		} else {
			sugar.Debug("mirror interval sync disabled")
		}

		if shouldEditRepo {
			sugar.Info("editing repo")
			_, _, err := client.EditRepo(owner, repoName, editRepoOption)
			if err != nil {
				reterr = errors.Join(reterr, fmt.Errorf("could not edit repo: %s: %w", teaRepo.FullName, err))
			}
		} else {
			sugar.Debug("not editing repo")
		}
	}

	if cfg.SyncTopics {
		sugar.Info("syncing topics")
		if teaTopics, _, err := client.ListRepoTopics(owner, repoName, gitea.ListRepoTopicsOptions{}); err != nil {
			reterr = errors.Join(reterr, fmt.Errorf("could not get repo topics: %s: %w", teaRepo.FullName, err))
		} else if isHubTopicsDifferent(teaTopics, hubRepo) {
			if _, err := client.SetRepoTopics(owner, repoName, hubRepo.Topics); err != nil {
				reterr = errors.Join(reterr, fmt.Errorf("could not set repo topics: %s: %w", teaRepo.FullName, err))
			}
		}
	} else {
		sugar.Debug("not syncing topics")
	}

	// Handle cases where the source repo had commits after it was archived
	if isHubRepoMirrorStale(teaRepo, hubRepo) {
		sugar.Info("mirror stale, updating", "src", hubRepo.PushedAt, "dest", teaRepo.MirrorUpdated)
		_, err := client.MirrorSync(owner, repoName)
		if err != nil {
			reterr = errors.Join(reterr, fmt.Errorf("could not mirror sync: %s: %w", teaRepo.FullName, err))
		}
	} else {
		sugar.Debugw("mirror not stale", "src", hubRepo.PushedAt, "dest", teaRepo.MirrorUpdated)
	}

	return reterr
}
