package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig holds database connection parameters
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// DatabasePair represents a source-target database pair to monitor
type DatabasePair struct {
	Name            string         `yaml:"name"`
	SourceDB        DatabaseConfig `yaml:"source_db"`
	TargetDB        DatabaseConfig `yaml:"target_db"`
	TablesToMonitor []string       `yaml:"tables_to_monitor"`
}

// Config holds the application configuration
type Config struct {
	// Legacy single database pair (for backward compatibility)
	SourceDB            DatabaseConfig   `yaml:"source_db,omitempty"`
	TargetDB            DatabaseConfig   `yaml:"target_db,omitempty"`
	TablesToMonitor     []string         `yaml:"tables_to_monitor,omitempty"`
	
	// New multi-database support
	DatabasePairs       []DatabasePair   `yaml:"database_pairs,omitempty"`
	
	MonitoringInterval  time.Duration    `yaml:"monitoring_interval"`
	ReplicaLagThreshold time.Duration    `yaml:"replica_lag_threshold"`
	WebServerPort       int              `yaml:"web_server_port"`
	LogLevel            string           `yaml:"log_level"`
}

// LoadConfig loads configuration from a YAML file with environment variable overrides
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Convert legacy single database config to database pairs format
	if config.SourceDB.Host != "" && len(config.DatabasePairs) == 0 {
		config.DatabasePairs = []DatabasePair{
			{
				Name:            "default",
				SourceDB:        config.SourceDB,
				TargetDB:        config.TargetDB,
				TablesToMonitor: config.TablesToMonitor,
			},
		}
	}

	// Apply environment variable overrides for legacy config
	if host := os.Getenv("SOURCE_DB_HOST"); host != "" {
		config.SourceDB.Host = host
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].SourceDB.Host = host
		}
	}
	if user := os.Getenv("SOURCE_DB_USERNAME"); user != "" {
		config.SourceDB.Username = user
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].SourceDB.Username = user
		}
	}
	if pass := os.Getenv("SOURCE_DB_PASSWORD"); pass != "" {
		config.SourceDB.Password = pass
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].SourceDB.Password = pass
		}
	}
	if host := os.Getenv("TARGET_DB_HOST"); host != "" {
		config.TargetDB.Host = host
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].TargetDB.Host = host
		}
	}
	if user := os.Getenv("TARGET_DB_USERNAME"); user != "" {
		config.TargetDB.Username = user
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].TargetDB.Username = user
		}
	}
	if pass := os.Getenv("TARGET_DB_PASSWORD"); pass != "" {
		config.TargetDB.Password = pass
		if len(config.DatabasePairs) > 0 {
			config.DatabasePairs[0].TargetDB.Password = pass
		}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.DatabasePairs) == 0 {
		return fmt.Errorf("at least one database pair must be configured")
	}

	// Validate each database pair
	for i, pair := range c.DatabasePairs {
		if pair.Name == "" {
			return fmt.Errorf("database pair %d: name is required", i)
		}

		// Validate source database
		if pair.SourceDB.Host == "" {
			return fmt.Errorf("database pair '%s': source database host is required", pair.Name)
		}
		if pair.SourceDB.Port == 0 {
			return fmt.Errorf("database pair '%s': source database port is required", pair.Name)
		}
		if pair.SourceDB.Username == "" {
			return fmt.Errorf("database pair '%s': source database username is required", pair.Name)
		}
		if pair.SourceDB.Database == "" {
			return fmt.Errorf("database pair '%s': source database name is required", pair.Name)
		}

		// Validate target database
		if pair.TargetDB.Host == "" {
			return fmt.Errorf("database pair '%s': target database host is required", pair.Name)
		}
		if pair.TargetDB.Port == 0 {
			return fmt.Errorf("database pair '%s': target database port is required", pair.Name)
		}
		if pair.TargetDB.Username == "" {
			return fmt.Errorf("database pair '%s': target database username is required", pair.Name)
		}
		if pair.TargetDB.Database == "" {
			return fmt.Errorf("database pair '%s': target database name is required", pair.Name)
		}
	}

	if c.MonitoringInterval < 10*time.Second {
		return fmt.Errorf("monitoring interval must be at least 10 seconds")
	}

	if c.WebServerPort == 0 {
		c.WebServerPort = 8080 // Default port
	}

	if c.ReplicaLagThreshold == 0 {
		c.ReplicaLagThreshold = 60 * time.Second // Default threshold
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	return nil
}
