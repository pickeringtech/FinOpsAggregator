package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Compute  ComputeConfig  `mapstructure:"compute"`
	Charts   ChartsConfig   `mapstructure:"charts"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Jobs     JobsConfig     `mapstructure:"jobs"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// PostgresConfig holds database configuration
type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

// ComputeConfig holds computation settings
type ComputeConfig struct {
	BaseCurrency     string   `mapstructure:"base_currency"`
	ActiveDimensions []string `mapstructure:"active_dimensions"`
}

// ChartsConfig holds chart generation settings
type ChartsConfig struct {
	OutDir string `mapstructure:"out_dir"`
}

// StorageConfig holds storage backend settings
type StorageConfig struct {
	URL    string `mapstructure:"url"`
	Prefix string `mapstructure:"prefix"`
}

// JobsConfig holds background job settings
type JobsConfig struct {
	Concurrency int            `mapstructure:"concurrency"`
	Queues      map[string]int `mapstructure:"queues"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// Load loads configuration from file and environment variables
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Environment variables
	v.SetEnvPrefix("FINOPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Postgres defaults
	v.SetDefault("postgres.dsn", "postgresql://finops:finops@localhost:5432/finops?sslmode=disable")

	// Compute defaults
	v.SetDefault("compute.base_currency", "USD")
	v.SetDefault("compute.active_dimensions", []string{
		"instance_hours",
		"storage_gb_month",
		"egress_gb",
		"iops",
		"backups_gb_month",
	})

	// Charts defaults
	v.SetDefault("charts.out_dir", "./charts")

	// Storage defaults
	v.SetDefault("storage.url", "file://./charts")
	v.SetDefault("storage.prefix", "")

	// Jobs defaults
	v.SetDefault("jobs.concurrency", 4)
	v.SetDefault("jobs.queues.default", 1)
	v.SetDefault("jobs.queues.exports", 1)

	// Logging defaults
	v.SetDefault("logging.level", "info")
}
