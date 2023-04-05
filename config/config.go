package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v7"
)

const DefaultDestMirrorInterval = "8h0m0s"

type Source string

const (
	SourceGitHub Source = "github"
	SourceGitea  Source = "gitea"
)

type Config struct {
	Daemon          int  `env:"DAEMON"`
	DaemonSkipFirst bool `env:"DAEMON_SKIP_FIRST"`
	DaemonExitError bool `env:"DAEMON_EXIT_ERROR"`

	Source      Source
	GitHubOwner string   `env:"GITHUB_OWNER"`
	GitHubToken string   `env:"GITHUB_TOKEN"`
	GiteaOwner  string   `env:"GITEA_OWNER"`
	GiteaToken  string   `env:"GITEA_TOKEN"`
	GiteaURL    string   `env:"GITEA_URL"`
	SkipRepos   []string `env:"SKIP_REPOS" envSeparator:" "`
	SkipForks   bool     `env:"SKIP_FORKS"`
	SkipPrivate bool     `env:"SKIP_PRIVATE"`

	MigrateAll  bool `env:"MIGRATE_ALL"`
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

func New() *Config {
	cfg := Config{}

	flag.IntVar(&cfg.Daemon, "daemon", 0, `Seconds between each run where 0 means running only once (e.g. "86400" is a day).`)
	flag.BoolVar(&cfg.DaemonSkipFirst, "daemon-skip-first", false, "Skip first run.")
	flag.BoolVar(&cfg.DaemonExitError, "daemon-exit-error", false, "Exit daemon when error occurs.")
	flag.StringVar(&cfg.GitHubOwner, "github-owner", "", "Owner of GitHub source repositories.")
	flag.StringVar(&cfg.GitHubToken, "github-token", "", "Token for accessing GitHub.")
	flag.StringVar(&cfg.GiteaOwner, "gitea-owner", "", "Owner of Gitea source repositories.")
	flag.StringVar(&cfg.GiteaToken, "gitea-token", "", "Token for accessing the source Gitea instance.")
	flag.StringVar(&cfg.GiteaURL, "gitea-url", "", "URL of the source Gitea instance.")
	skipRepos := flag.String("skip-repos", "", `List of space seperated repositories to not sync (e.g. "ItsNotGoodName/example1 itsnotgoodname/example2 example3").`)
	flag.BoolVar(&cfg.SkipForks, "skip-forks", false, "Skip fork repositories.")
	flag.BoolVar(&cfg.SkipPrivate, "skip-private", false, "Skip private repositories.")
	flag.BoolVar(&cfg.MigrateAll, "migrate-all", false, "Migrate everything.")
	flag.BoolVar(&cfg.MigrateWiki, "migrate-wiki", false, "Migrate wiki from source repositories.")
	flag.BoolVar(&cfg.SyncAll, "sync-all", false, "Sync everything.")
	flag.BoolVar(&cfg.SyncTopics, "sync-topics", false, "Sync topics of repository.")
	flag.BoolVar(&cfg.SyncDescription, "sync-description", false, "Sync description of repository.")
	flag.BoolVar(&cfg.SyncVisibility, "sync-visibility", false, "Sync private/public status of repository.")
	flag.BoolVar(&cfg.SyncMirrorInterval, "sync-mirror-interval", false, "Disable periodic sync if source repository is archived.")
	flag.StringVar(&cfg.DestURL, "dest-url", "", "URL of the destination Gitea instance. (required)")
	flag.StringVar(&cfg.DestToken, "dest-token", "", "Token for accessing the destination Gitea instance. (required)")
	flag.StringVar(&cfg.DestOwner, "dest-owner", "", "Owner of the mirrored repositories on the destination Gitea instance.")
	flag.StringVar(&cfg.DestMirrorInterval, "dest-mirror-interval", DefaultDestMirrorInterval, "Default mirror interval for new migrations on the destination Gitea instance.")

	flag.Parse()

	cfg.SkipRepos = strings.Split(*skipRepos, " ")

	return &cfg
}

func (cfg *Config) ParseAndValidate() error {
	if err := env.Parse(cfg); err != nil {
		return err
	}

	if cfg.MigrateAll {
		cfg.MigrateWiki = true
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

	if cfg.Daemon < 60 && cfg.Daemon != 0 {
		return fmt.Errorf("DAEMON interval too quick: %d", cfg.Daemon)
	}

	return nil
}
