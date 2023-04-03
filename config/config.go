package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v7"
)

// type Source string

// const (
// 	SourceGithub Source = "github"
// )

type Config struct {
	// SrcType        Source   `env:"SRC_TYPE" envDefault:"github"`
	// SrcURL         string   `env:"SRC_URL"`
	SrcToken string `env:"SRC_TOKEN"`
	SrcOwner string `env:"SRC_OWNER"`
	// SrcRepos       []string `env:"SRC_REPOS" envSeparator:" "`
	// SrcSkip        []string `env:"SRC_SKIP" envSeparator:" "`
	SrcSkipForks   bool `env:"SRC_SKIP_FORKS"`
	SrcSkipPrivate bool `env:"SRC_SKIP_PRIVATE"`

	SyncAll            bool `env:"SYNC_ALL"`
	SyncTopics         bool `env:"SYNC_TOPICS"`
	SyncDescription    bool `env:"SYNC_DESCRIPTION"`
	SyncVisibility     bool `env:"SYNC_VISIBILITY"`
	SyncMirrorInterval bool `env:"SYNC_MIRROR_INTERVAL"`

	DestURL            string `env:"DEST_URL"`
	DestToken          string `env:"DEST_TOKEN"`
	DestOwner          string `env:"DEST_OWNER"`
	DestMirrorInterval string `env:"DEST_MIRROR_INTERVAL" envDefault:"8h0m0s"`
}

const DefaultMirrorInterval = "8h0m0s"

func New() *Config {
	return &Config{}
}

func (cfg *Config) WithFlags() *Config {
	flag.StringVar(&cfg.SrcToken, "src-token", "", "Token for source service.")
	flag.StringVar(&cfg.SrcOwner, "src-owner", "", "Owner of source repositories to mirror.")
	flag.BoolVar(&cfg.SrcSkipForks, "src-skip-forks", false, "Skip source repositories that are forks.")
	flag.BoolVar(&cfg.SrcSkipPrivate, "src-skip-private", false, "Skip source repositories that are private.")
	flag.BoolVar(&cfg.SyncAll, "sync-all", false, "Synchronize everything.")
	flag.BoolVar(&cfg.SyncTopics, "sync-topics", false, "Synchronize topics.")
	flag.BoolVar(&cfg.SyncDescription, "sync-description", false, "Synchronize description.")
	flag.BoolVar(&cfg.SyncVisibility, "sync-visibility", false, "Synchronize visibility status.")
	flag.BoolVar(&cfg.SyncMirrorInterval, "sync-mirror-interval", false, "Disable periodic sync if source repository is archived.")
	flag.StringVar(&cfg.DestURL, "dest-url", "", "URL of the destination Gitea instance. (required)")
	flag.StringVar(&cfg.DestToken, "dest-token", "", "Token for the destination Gitea instance. (required)")
	flag.StringVar(&cfg.DestOwner, "dest-owner", "", "Owner of the mirrors on the Gitea instance.")
	flag.StringVar(&cfg.DestMirrorInterval, "dest-mirror-interval", DefaultMirrorInterval, "Default mirror interval for new migrations on the Gitea instance.")

	flag.Parse()

	return cfg
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

	if cfg.SrcOwner == "" && cfg.SrcToken == "" {
		return fmt.Errorf("SRC_OWNER or SRC_TOKEN not set")
	}

	if cfg.DestURL == "" {
		return fmt.Errorf("DEST_URL not set")
	}

	if cfg.DestToken == "" {
		return fmt.Errorf("DEST_TOKEN not set")
	}

	return nil
}
