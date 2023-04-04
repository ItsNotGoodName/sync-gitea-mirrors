package main

import (
	"context"
	"fmt"

	"github.com/ItsNotGoodName/sync-gitea-mirrors/config"
	"github.com/ItsNotGoodName/sync-gitea-mirrors/hub"
	"github.com/ItsNotGoodName/sync-gitea-mirrors/tea"
	"go.uber.org/zap"

	"code.gitea.io/sdk/gitea"
)

var log *zap.Logger

func main() {
	log, _ = zap.NewProduction()
	defer log.Sync()

	// Parse config
	cfg := config.New()
	if err := cfg.ParseAndValidate(); err != nil {
		log.Fatal("could not parse config", zap.Error(err))
	}

	if cfg.Source == config.SourceGitHub {
		if cfg.GitHubOwner != "" && cfg.GitHubToken != "" {
			log.Warn("settings both GITHUB_OWNER and GITHUB_TOKEN will only display public repos")
		}
	}

	syncConfig := tea.SyncConfig{
		SyncDescription:    cfg.SyncDescription,
		SyncMirrorInterval: cfg.SyncMirrorInterval,
		SyncTopics:         cfg.SyncTopics,
		SyncVisibility:     cfg.SyncVisibility,
		DestMirrorInterval: cfg.DestMirrorInterval,
	}

	// Create client
	client, err := gitea.NewClient(cfg.DestURL, gitea.SetToken(cfg.DestToken))
	if err != nil {
		log.Fatal("could not create dest Gitea client", zap.Error(err))
	}

	// Get repositories based config
	repos, migrateRepoOption := getSourceRepos(cfg)

	fmt.Printf("Will sync %d repositories\n", len(repos))

Loop:
	for _, repo := range repos {
		for _, skip := range cfg.Skip {
			if repo.Is(skip) {
				fmt.Println("Skipping", repo.GetFullName())
				continue Loop
			}
		}

		owner := cfg.DestOwner
		if owner == "" {
			owner = repo.Owner
		}
		name := repo.Name

		teaRepo, err := tea.GetRepo(client, owner, name)
		if err != nil {
			log.Error("could not get dest repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
			continue
		}

		// Migrate new repo
		if teaRepo == nil {
			fmt.Println("Migrating", repo.GetFullName())

			opts := migrateRepoOption
			opts.Mirror = true
			opts.RepoOwner = owner
			opts.RepoName = name
			opts.Private = repo.Private
			opts.Wiki = cfg.MigrateWiki

			if teaRepo, _, err = client.MigrateRepo(opts); err != nil {
				log.Error("could not migrate repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
				continue
			}
		}

		// Sync existing repo
		fmt.Println("Syncing", repo.GetFullName())
		output, err := tea.Sync(client, teaRepo, &repo.SyncRepository, syncConfig)
		if err != nil {
			log.Error("could not sync repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
		}

		if output.SyncMirror {
			fmt.Println("~ Synced mirror")
		}
		if output.UpdateDescription {
			fmt.Println("~ Updated description")
		}
		if output.UpdateMirrorInterval {
			fmt.Println("~ Updated mirror-interval")
		}
		if output.UpdateTopics {
			fmt.Println("~ Updated topics")
		}
		if output.UpdateVisibility {
			fmt.Println("~ Updated visibility")
		}
	}
}

func getSourceRepos(cfg *config.Config) ([]tea.SourceRepository, gitea.MigrateRepoOption) {
	switch cfg.Source {
	case config.SourceGitHub:
		// Create GitHub client
		ctx := context.Background()
		hubClient := hub.NewClient(ctx, cfg.GitHubToken)

		// List repositories
		repos, err := hub.ListRepos(ctx, hubClient, cfg.GitHubOwner, cfg.SkipPrivate, cfg.SkipForks)
		if err != nil {
			log.Fatal("could not get GitHub repos", zap.Error(err))
		}

		return hub.ConvertList(repos), gitea.MigrateRepoOption{
			Service:   gitea.GitServiceGithub,
			AuthToken: cfg.GitHubToken,
		}
	case config.SourceGitea:
		// Create Gitea client
		srcClient, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
		if err != nil {
			log.Fatal("could not create source Gitea client", zap.Error(err))
		}

		// List repositories
		repos, err := tea.ListRepos(srcClient, cfg.GiteaOwner, cfg.SkipPrivate, cfg.SkipForks)
		if err != nil {
			log.Fatal("could not set source Gitea repos", zap.Error(err))
		}

		var getTopics func(r *gitea.Repository) []string
		if cfg.SyncTopics {
			getTopics = func(r *gitea.Repository) []string {
				topics, _, err := srcClient.ListRepoTopics(r.Owner.UserName, r.Name, gitea.ListRepoTopicsOptions{})
				if err != nil {
					log.Fatal("could not list topics", zap.String("repo", r.FullName), zap.Error(err))
				}
				return topics
			}
		} else {
			getTopics = func(r *gitea.Repository) []string {
				return []string{}
			}
		}

		return tea.ConvertList(repos, getTopics), gitea.MigrateRepoOption{
			Service:   gitea.GitServiceGitea,
			AuthToken: cfg.GiteaToken,
		}
	default:
		panic(fmt.Sprintf("invalid SOURCE: %s", cfg.Source))
	}
}
