// Package config provides configuration management for the AI Gateway.
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	App       AppConfig        `yaml:"app"`
	Log       LogConfig        `yaml:"log"`
	HTTP      HTTPConfig       `yaml:"http"`
	MySQL     MySQLConfig      `yaml:"mysql"`
	Auth      AuthConfig       `yaml:"auth"`
	Providers []ProviderConfig `yaml:"providers"`
	Models    ModelsConfig     `yaml:"models"`
}

// AppConfig contains application-level settings.
type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"` // development, staging, production
}

// LogConfig contains logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, console
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

// MySQLConfig contains MySQL database settings.
type MySQLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"`
	MaxOpen  int    `yaml:"maxOpen"`
}

// ProviderConfig contains settings for a single provider instance.
type ProviderConfig struct {
	Name    string        `yaml:"name"` // Unique identifier
	Type    string        `yaml:"type"` // "openai" | "anthropic"
	APIKey  string        `yaml:"apiKey"`
	BaseURL string        `yaml:"baseURL"`
	Timeout time.Duration `yaml:"timeout"`
	Default bool          `yaml:"default"` // Default provider for this type
}

// ModelsConfig defines model routing and load balancing.
type ModelsConfig struct {
	// Routing contains exact model -> provider mappings
	Routing map[string]ModelRoute `yaml:"routing"`
	// PrefixRouting contains prefix-based routing rules
	PrefixRouting map[string]PrefixRoute `yaml:"prefixRouting"`
	// LoadBalancing contains load balancing configurations
	LoadBalancing map[string]LoadBalanceConfig `yaml:"loadBalancing"`
}

// ModelRoute defines where a model request should be routed (exact match).
type ModelRoute struct {
	Provider    string `yaml:"provider"`    // Provider name
	ActualModel string `yaml:"actualModel"` // Optional: actual model name to use
}

// PrefixRoute defines prefix-based routing (e.g., "deepseek-" -> siliconflow).
type PrefixRoute struct {
	Provider string `yaml:"provider"`
	Priority int    `yaml:"priority"` // Higher priority wins when multiple prefixes match
}

// LoadBalanceConfig defines load balancing for a model pattern.
type LoadBalanceConfig struct {
	Strategy  string              `yaml:"strategy"` // "round-robin" | "random" | "failover" | "weighted"
	Providers []LoadBalanceMember `yaml:"providers"`
}

// LoadBalanceMember represents a provider in a load balance group.
type LoadBalanceMember struct {
	Name     string `yaml:"name"`     // Provider name
	Weight   int    `yaml:"weight"`   // For weighted strategy
	Priority int    `yaml:"priority"` // For failover strategy (lower = higher priority)
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	Enabled bool     `yaml:"enabled"`
	APIKeys []string `yaml:"apiKeys"`
}

// Load reads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.HTTP.ReadTimeout == 0 {
		cfg.HTTP.ReadTimeout = 30 * time.Second
	}
	if cfg.HTTP.WriteTimeout == 0 {
		cfg.HTTP.WriteTimeout = 120 * time.Second
	}

	// Set default timeout for providers
	for i := range cfg.Providers {
		if cfg.Providers[i].Timeout == 0 {
			cfg.Providers[i].Timeout = 60 * time.Second
		}
	}

	// Initialize maps if nil
	if cfg.Models.Routing == nil {
		cfg.Models.Routing = make(map[string]ModelRoute)
	}
	if cfg.Models.PrefixRouting == nil {
		cfg.Models.PrefixRouting = make(map[string]PrefixRoute)
	}
	if cfg.Models.LoadBalancing == nil {
		cfg.Models.LoadBalancing = make(map[string]LoadBalanceConfig)
	}

	// Set default weight for load balance members
	for model, lb := range cfg.Models.LoadBalancing {
		for i := range lb.Providers {
			if lb.Providers[i].Weight == 0 {
				lb.Providers[i].Weight = 1
			}
		}
		cfg.Models.LoadBalancing[model] = lb
	}

	return &cfg, nil
}

// DefaultConfig returns a default configuration for development.
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name: "ai-gateway",
			Env:  "development",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "console",
		},
		HTTP: HTTPConfig{
			Addr:         ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
		},
		Providers: []ProviderConfig{},
		Models: ModelsConfig{
			Routing:       make(map[string]ModelRoute),
			PrefixRouting: make(map[string]PrefixRoute),
			LoadBalancing: make(map[string]LoadBalanceConfig),
		},
	}
}
