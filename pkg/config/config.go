package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	CFURL              string   `mapstructure:"cf_url"`
	CFDomain           string   `mapstructure:"cf_domain"`
	CFNames            []string `mapstructure:"cf_names"`
	GcoreURL           string   `mapstructure:"gcore_url"`
	GcoreDomain        string   `mapstructure:"gcore_domain"`
	GcoreNames         []string `mapstructure:"gcore_names"`
	CloudflareAPIToken string   `mapstructure:"cloudflare_api_token"`
	MaxDelay           int      `mapstructure:"max_delay"`
}

func Load(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.CFURL == "" {
		return errors.New("CF_URL is required")
	}
	if cfg.CFDomain == "" {
		return errors.New("CF_DOMAIN is required")
	}
	if len(cfg.CFNames) == 0 {
		return errors.New("CF_NAMES is required")
	}
	if cfg.GcoreURL == "" {
		return errors.New("GCORE_URL is required")
	}
	if cfg.GcoreDomain == "" {
		return errors.New("GCORE_DOMAIN is required")
	}
	if len(cfg.GcoreNames) == 0 {
		return errors.New("GCORE_NAMES is required")
	}
	if cfg.CloudflareAPIToken == "" {
		return errors.New("CLOUDFLARE_API_TOKEN is required")
	}
	if cfg.MaxDelay <= 0 {
		return errors.New("MAX_DELAY must be greater than 0")
	}
	return nil
}