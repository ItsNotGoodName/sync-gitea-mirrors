package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Source string

const (
	SourceGithub Source = "github"
)

type Config struct {
	SrcType        Source   `env:"SRC_TYPE" envDefault:"github"`
	SrcURL         string   `env:"SRC_URL"`
	SrcToken       string   `env:"SRC_TOKEN"`
	SrcOwner       string   `env:"SRC_OWNER"`
	SrcRepos       []string `env:"SRC_REPOS" envSeparator:" "`
	SrcSkip        []string `env:"SRC_SKIP" envSeparator:" "`
	SrcSkipForks   bool     `env:"SRC_SKIP_FORKS"`
	SrcSkipPrivate bool     `env:"SRC_SKIP_PRIVATE"`

	SyncTopics         bool `env:"SYNC_TOPICS"`
	SyncDescription    bool `env:"SYNC_DESCRIPTION"`
	SyncVisibility     bool `env:"SYNC_VISIBILITY"`
	SyncMirrorInterval bool `env:"SYNC_MIRROR_INTERVAL"`

	DestURL            string `env:"DEST_URL"`
	DestToken          string `env:"DEST_TOKEN"`
	DestOwner          string `env:"DEST_OWNER"`
	DestMirrorInterval string `env:"DEST_MIRROR_INTERVAL" envDefault:"8h0m0s"`
}

func Parse(cfg *Config) error {
	if err := env.Parse(cfg); err != nil {
		return err
	}

	if cfg.SrcOwner == "" && cfg.SrcToken == "" {
		return fmt.Errorf("SRC_OWNER or SRC_TOKEN not set")
	}

	if cfg.SrcOwner != "" && cfg.SrcToken != "" {
		return fmt.Errorf("SRC_OWNER and SRC_TOKEN cannot both be set")
	}

	if cfg.DestURL == "" {
		return fmt.Errorf("DEST_URL not set")
	}

	if cfg.DestToken == "" {
		return fmt.Errorf("DEST_TOKEN not set")
	}

	return nil
}
