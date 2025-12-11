package githubbackup

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	BucketName  string   `env:"BUCKET_NAME, required"`
	GitHubToken string   `env:"GITHUB_TOKEN, required"`
	GitHubOrgs  []string `env:"GITHUB_ORGS, default=navikt,nais"`
}

func newConfig(ctx context.Context, lookuper envconfig.Lookuper) (*Config, error) {
	cfg := &Config{}
	err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   cfg,
		Lookuper: lookuper,
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
