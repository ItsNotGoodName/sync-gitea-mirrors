package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ItsNotGoodName/sync-gitea-mirrors/config"
	"github.com/ItsNotGoodName/sync-gitea-mirrors/hub"
	"github.com/ItsNotGoodName/sync-gitea-mirrors/tea"
	"go.uber.org/zap"

	"code.gitea.io/sdk/gitea"
)

var log *zap.Logger
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log, _ = zap.NewProduction()
	defer log.Sync()

	// Read flags
	cfg := config.New()

	// Show version and exit
	if cfg.ShowVersion {
		fmt.Println(version)
		return
	}

	// Show info and exit
	if cfg.ShowInfo {
		fmt.Printf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date)
		return
	}

	// Parse config
	if err := cfg.ParseAndValidate(); err != nil {
		log.Fatal("could not parse config", zap.Error(err))
	}

	if cfg.Source == config.SourceGitHub {
		if cfg.GitHubOwner != "" {
			log.Warn("setting GITHUB_OWNER will only show public repositories")
		}
	}

	syncConfig := tea.SyncConfig{
		SyncDescription:    cfg.SyncDescription,
		SyncMirrorInterval: cfg.SyncMirrorInterval,
		SyncTopics:         cfg.SyncTopics,
		SyncVisibility:     cfg.SyncVisibility,
		DestMirrorInterval: cfg.DestMirrorInterval,
	}

	fmt.Printf("SyncConfig: %+v\n", syncConfig)

	if cfg.Daemon != 0 {
		// Daemon
		interval := time.Duration(cfg.Daemon) * time.Second

		if cfg.DaemonSkipFirst {
			fmt.Println("Sleeping for", cfg.Daemon, "seconds")
			time.Sleep(interval)
		}

		for {
			if err := run(cfg, &syncConfig); err != nil {
				if cfg.DaemonExitError {
					log.Fatal("main", zap.Error(err))
				}
				log.Error("main", zap.Error(err))
			}

			fmt.Println("Sleeping for", cfg.Daemon, "seconds")
			time.Sleep(interval)
		}
	} else {
		// Normal
		if err := run(cfg, &syncConfig); err != nil {
			log.Fatal("main", zap.Error(err))
		}
	}
}

func run(cfg *config.Config, syncConfig *tea.SyncConfig) error {
	// Create client
	client, err := gitea.NewClient(cfg.DestURL, gitea.SetToken(cfg.DestToken))
	if err != nil {
		return fmt.Errorf("could not create destination Gitea client: %w", err)
	}

	// Get repositories based config
	repos, migrateRepoOption, err := getSourceRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Will sync %d repositories\n", len(repos))

Loop:
	for _, repo := range repos {
		// Skip
		for _, skipRepo := range cfg.SkipRepos {
			if repo.Is(skipRepo) {
				fmt.Println("Skipping", repo.GetFullName())
				continue Loop
			}
		}

		// Destination repo name and owner
		owner := cfg.DestOwner
		if owner == "" {
			owner = repo.Owner
		}
		name := repo.Name

		teaRepo, err := tea.GetRepoOrNil(client, owner, name)
		if err != nil {
			log.Error("could not get destination repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
			continue
		}

		// Migrate new repo
		if teaRepo == nil {
			fmt.Println("Migrating", repo.GetFullName())

			opts := migrateRepoOption
			opts.Mirror = true
			opts.RepoOwner = owner
			opts.RepoName = name
			opts.CloneAddr = repo.URLS[0]
			opts.Private = repo.Private
			opts.Wiki = cfg.MigrateWiki
			opts.LFS = cfg.MigrateLFS

			if teaRepo, _, err = client.MigrateRepo(opts); err != nil {
				log.Error("could not migrate repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
				continue
			}
		} else if !repo.IsMyMirror(teaRepo) {
			fmt.Println("Skipping", repo.GetFullName(), "does not belong to mirror", teaRepo.FullName)
			continue
		}

		// Sync existing repo
		fmt.Println("Syncing", repo.GetFullName())
		output, err := tea.Sync(client, teaRepo, &repo.SyncRepository, syncConfig)
		if err != nil {
			log.Error("could not sync repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
			continue
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

	return nil
}

func getSourceRepos(cfg *config.Config) ([]tea.SourceRepository, gitea.MigrateRepoOption, error) {
	switch cfg.Source {
	case config.SourceGitHub:
		// Create GitHub client
		ctx := context.Background()
		hubClient := hub.NewClient(ctx, cfg.GitHubToken)

		// List repositories
		repos, err := hub.ListRepos(ctx, hubClient, cfg.GitHubOwner, cfg.SkipPrivate, cfg.SkipForks)
		if err != nil {
			return nil, gitea.MigrateRepoOption{}, fmt.Errorf("could not get GitHub repos: %s: %w", cfg.GitHubOwner, err)
		}

		return hub.ConvertList(repos), gitea.MigrateRepoOption{
			Service:   gitea.GitServiceGithub,
			AuthToken: cfg.GitHubToken,
		}, nil
	case config.SourceGitea:
		// Create Gitea client
		srcClient, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
		if err != nil {
			return nil, gitea.MigrateRepoOption{}, fmt.Errorf("could not create source Gitea client: %s: %w", cfg.GiteaURL, err)
		}

		// List repositories
		repos, err := tea.ListRepos(srcClient, cfg.GiteaOwner, cfg.SkipPrivate, cfg.SkipForks)
		if err != nil {
			return nil, gitea.MigrateRepoOption{}, fmt.Errorf("could not set source Gitea repos: %s: %w", cfg.GiteaOwner, err)
		}

		var getTopics func(r *gitea.Repository) ([]string, error)
		if cfg.SyncTopics {
			getTopics = func(r *gitea.Repository) ([]string, error) {
				topics, _, err := srcClient.ListRepoTopics(r.Owner.UserName, r.Name, gitea.ListRepoTopicsOptions{})
				if err != nil {
					return nil, fmt.Errorf("could not list topics: %s: %w", r.FullName, err)
				}

				return topics, nil
			}
		} else {
			getTopics = func(r *gitea.Repository) ([]string, error) {
				return []string{}, nil
			}
		}

		convRepos, err := tea.ConvertList(repos, getTopics)
		if err != nil {
			return nil, gitea.MigrateRepoOption{}, err
		}

		return convRepos, gitea.MigrateRepoOption{
			Service:   gitea.GitServiceGitea,
			AuthToken: cfg.GiteaToken,
		}, nil
	default:
		panic(fmt.Sprintf("invalid SOURCE: %s", cfg.Source))
	}
}
