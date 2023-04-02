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

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	// Parse config
	cfg := config.New().WithFlags()
	if err := cfg.ParseAndValidate(); err != nil {
		log.Fatal("could not parse config", zap.Error(err))
	}

	// Create clients
	ctx := context.Background()
	hubClient := hub.NewClient(ctx, cfg.SrcToken)
	client, err := gitea.NewClient(cfg.DestURL, gitea.SetToken(cfg.DestToken))
	if err != nil {
		log.Fatal("could not create Gitea client", zap.Error(err))
	}

	repos, err := hub.GetRepos(ctx, hubClient, cfg.SrcOwner)
	if err != nil {
		log.Fatal("could not get GitHub repos", zap.Error(err))
	}

	fmt.Printf("Will sync %d repositories\n", len(repos))

	syncConfig := tea.SyncConfig{
		SyncDescription:    cfg.SyncDescription,
		SyncMirrorInterval: cfg.SyncMirrorInterval,
		SyncTopics:         cfg.SyncTopics,
		SyncVisibility:     cfg.SyncVisibility,
		DestMirrorInterval: cfg.DestMirrorInterval,
	}

	for _, repo := range repos {
		owner := cfg.DestOwner
		if owner == "" {
			owner = repo.Owner.GetLogin()
		}
		name := repo.GetName()

		teaRepo, err := tea.GetRepo(client, owner, name)
		if err != nil {
			log.Error("could not get dest repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
			continue
		}

		// Migrate new repo
		if teaRepo == nil {
			fmt.Println("Migrating", repo.GetFullName())
			if err = hub.Migrate(client, gitea.MigrateRepoOption{
				Mirror:    true,
				RepoOwner: owner,
				RepoName:  name,
			}, repo, cfg.SrcToken); err != nil {
				log.Error("could not migrate repo", zap.String("owner", owner), zap.String("name", name), zap.Error(err))
			}
			continue
		}

		// Sync existing repo
		fmt.Println("Syncing", repo.GetFullName())
		output, err := tea.Sync(client, teaRepo, &tea.SourceRepository{
			Topics:      repo.Topics,
			Description: repo.GetDescription(),
			Private:     repo.GetPrivate(),
			Archived:    repo.GetArchived(),
			PushedAt:    repo.GetPushedAt().Time,
		}, syncConfig)
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
