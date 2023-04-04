package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v7"
)

type Source string

const (
	SourceGitHub Source = "github"
	SourceGitea  Source = "gitea"
)

type Config struct {
	Source      Source
	GitHubOwner string   `env:"GITHUB_OWNER"`
	GitHubToken string   `env:"GITHUB_TOKEN"`
	GiteaOwner  string   `env:"GITEA_OWNER"`
	GiteaToken  string   `env:"GITEA_TOKEN"`
	GiteaURL    string   `env:"GITEA_URL"`
	Skip        []string `env:"SKIP" envSeparator:" "`
	SkipForks   bool     `env:"SKIP_FORKS"`
	SkipPrivate bool     `env:"SKIP_PRIVATE"`

	MigrateWiki bool `env:"MIGRATE_WIKI"`

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

	flag.StringVar(&cfg.GitHubOwner, "github-owner", "", "Owner of GitHub repositories to mirror.")
	flag.StringVar(&cfg.GitHubToken, "github-token", "", "GitHub token for mirroring and syncing.")
	flag.StringVar(&cfg.GitHubOwner, "gitea-owner", "", "Owner of Gitea repositories to mirror.")
	flag.StringVar(&cfg.GiteaToken, "gitea-token", "", "Gitea token for mirroring and syncing.")
	flag.StringVar(&cfg.GiteaURL, "gitea-url", "", "URL for the source Gitea instance.")
	skip := flag.String("skip", "", `List of source repositories to skip seperated by " " (e.g. "ItsNotGoodName/example1 itsnotgoodname/example2 example3").`)
	flag.BoolVar(&cfg.SkipForks, "skip-forks", false, "Skip source repositories that are forks.")
	flag.BoolVar(&cfg.SkipPrivate, "skip-private", false, "Skip source repositories that are private.")
	flag.BoolVar(&cfg.MigrateWiki, "migrate-wiki", false, "Migrate wiki.")
	flag.BoolVar(&cfg.SyncAll, "sync-all", false, "Synchronize everything.")
	flag.BoolVar(&cfg.SyncTopics, "sync-topics", false, "Synchronize repository topics.")
	flag.BoolVar(&cfg.SyncDescription, "sync-description", false, "Synchronize repository description.")
	flag.BoolVar(&cfg.SyncVisibility, "sync-visibility", false, "Synchronize repository visibility.")
	flag.BoolVar(&cfg.SyncMirrorInterval, "sync-mirror-interval", false, "Disable periodic sync if source repository is archived.")
	flag.StringVar(&cfg.DestURL, "dest-url", "", "URL of the destination Gitea instance.")
	flag.StringVar(&cfg.DestToken, "dest-token", "", "Token for the destination Gitea instance.")
	flag.StringVar(&cfg.DestOwner, "dest-owner", "", "Owner of the mirrors on the destination Gitea instance.")
	flag.StringVar(&cfg.DestMirrorInterval, "dest-mirror-interval", DefaultMirrorInterval, "Default mirror interval for new migrations on the destination Gitea instance.")

	flag.Parse()

	cfg.Skip = strings.Split(*skip, " ")

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

	// Infer source
	if cfg.GitHubOwner != "" || cfg.GitHubToken != "" {
		cfg.Source = SourceGitHub
	} else if cfg.GiteaOwner != "" || cfg.GiteaToken != "" || cfg.GiteaURL != "" {
		cfg.Source = SourceGitea
	} else {
		return fmt.Errorf("setup GitHub or Gitea as a repository source")
	}

	// Validate source config
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
		return fmt.Errorf("invalid SOURCE: %s", cfg.Source)
	}

	if cfg.DestURL == "" {
		return fmt.Errorf("DEST_URL not set")
	}

	if cfg.DestToken == "" {
		return fmt.Errorf("DEST_TOKEN not set")
	}

	return nil
}
