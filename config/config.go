package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Source string

const (
	SourceGitHub Source = "github"
	SourceGitea  Source = "gitea"
)

type Config struct {
	// SrcURL         string   `env:"SRC_URL"`
	// SrcRepos       []string `env:"SRC_REPOS" envSeparator:" "`
	// SrcSkip        []string `env:"SRC_SKIP" envSeparator:" "`

	Source      Source `env:"SRC"`
	GitHubOwner string `env:"GITHUB_OWNER"`
	GitHubToken string `env:"GITHUB_TOKEN"`
	GiteaOwner  string `env:"GITEA_OWNER"`
	GiteaToken  string `env:"GITEA_TOKEN"`
	GiteaURL    string `env:"GITEA_URL"`
	SkipForks   bool   `env:"SKIP_FORKS"`
	SkipPrivate bool   `env:"SKIP_PRIVATE"`

	SyncAll            bool `env:"SYNC_ALL"`
	SyncTopics         bool `env:"SYNC_TOPICS"`
	SyncDescription    bool `env:"SYNC_DESCRIPTION"`
	SyncVisibility     bool `env:"SYNC_VISIBILITY"`
	SyncMirrorInterval bool `env:"SYNC_MIRROR_INTERVAL"`

	DestURL            string `env:"DEST_URL"`
	DestToken          string `env:"DEST_TOKEN"`
	DestOwner          string `env:"DEST_OWNER"`
	DestMirrorInterval string `env:"DEST_MIRROR_INTERVAL"`
}

const DefaultMirrorInterval = "8h0m0s"

func New() *Config {
	cfg := Config{}

	flag.StringVar((*string)(&cfg.Source), "src", string(SourceGitHub), "Source service.")
	flag.StringVar(&cfg.GitHubOwner, "github-owner", "", "Owner of GitHub repositories to mirror.")
	flag.StringVar(&cfg.GitHubToken, "github-token", "", "Token for GitHub for mirroring and syncing.")
	flag.StringVar(&cfg.GitHubOwner, "gitea-owner", "", "Owner of Gitea repositories to mirror.")
	flag.StringVar(&cfg.GiteaToken, "gitea-token", "", "Token for Gitea for mirroring and syncing.")
	flag.StringVar(&cfg.GiteaURL, "gitea-url", "", "URL for the source Gitea instance.")
	flag.BoolVar(&cfg.SkipForks, "skip-forks", false, "Skip source repositories that are forks.")
	flag.BoolVar(&cfg.SkipPrivate, "skip-private", false, "Skip source repositories that are private.")
	flag.BoolVar(&cfg.SyncAll, "sync-all", false, "Synchronize everything.")
	flag.BoolVar(&cfg.SyncTopics, "sync-topics", false, "Synchronize repository topics.")
	flag.BoolVar(&cfg.SyncDescription, "sync-description", false, "Synchronize repository description.")
	flag.BoolVar(&cfg.SyncVisibility, "sync-visibility", false, "Synchronize repository visibility.")
	flag.BoolVar(&cfg.SyncMirrorInterval, "sync-mirror-interval", false, "Disable periodic sync if source repository is archived.")
	flag.StringVar(&cfg.DestURL, "dest-url", "", "URL of the destination Gitea instance.")
	flag.StringVar(&cfg.DestToken, "dest-token", "", "Token for the destination Gitea instance.")
	flag.StringVar(&cfg.DestOwner, "dest-owner", "", "Owner of the mirrors on the Gitea instance.")
	flag.StringVar(&cfg.DestMirrorInterval, "dest-mirror-interval", DefaultMirrorInterval, "Default mirror interval for new migrations on the Gitea instance.")

	flag.Parse()

	return &cfg
}

func (cfg *Config) ParseAndValidate() error {
	if err := env.Parse(cfg); err != nil {
		return err
	}

	if cfg.SyncAll {
		cfg.SyncDescription = true
		cfg.SyncMirrorInterval = true
		cfg.SyncTopics = true
		cfg.SyncVisibility = true
	}

	switch cfg.Source {
	case SourceGitHub:
		if cfg.GitHubOwner == "" && cfg.GitHubToken == "" {
			return fmt.Errorf("GITHUB_OWNER or GITHUB_TOKEN not set")
		}
	case SourceGitea:
		if cfg.GiteaOwner == "" && cfg.GiteaToken == "" {
			return fmt.Errorf("GITEA_OWNER or GITEA_TOKEN not set")
		}
		if cfg.GiteaURL == "" {
			return fmt.Errorf("GITEA_URL not set")
		}
	default:
		return fmt.Errorf("invalid SRC: %s", cfg.Source)
	}

	if cfg.DestURL == "" {
		return fmt.Errorf("DEST_URL not set")
	}

	if cfg.DestToken == "" {
		return fmt.Errorf("DEST_TOKEN not set")
	}

	return nil
}
