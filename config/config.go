// Package config 为 AI 网关提供配置管理。
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 代表应用程序配置。
type Config struct {
	App       AppConfig        `yaml:"app"`
	Log       LogConfig        `yaml:"log"`
	HTTP      HTTPConfig       `yaml:"http"`
	MySQL     MySQLConfig      `yaml:"mysql"`
	Auth      AuthConfig       `yaml:"auth"`
	Providers []ProviderConfig `yaml:"providers"`
	Models    ModelsConfig     `yaml:"models"`
}

// AppConfig 包含应用程序级别的设置。
type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"` // development, staging, production
}

// LogConfig 包含日志设置。
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, console
}

// HTTPConfig 包含 HTTP 服务器设置。
type HTTPConfig struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

// MySQLConfig 包含 MySQL 数据库设置。
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"`
	MaxOpen  int    `yaml:"maxOpen"`
}

// ProviderConfig 包含单个供应商实例的设置。
type ProviderConfig struct {
	Name    string        `yaml:"name"` // 唯一标识符
	Type    string        `yaml:"type"` // "openai" | "anthropic"
	APIKey  string        `yaml:"apiKey"`
	BaseURL string        `yaml:"baseURL"`
	Timeout time.Duration `yaml:"timeout"`
	Default bool          `yaml:"default"` // 此类型的默认供应商
}

// ModelsConfig 定义模型路由和负载均衡。
type ModelsConfig struct {
	// Routing 包含精确的模型 -> 供应商映射
	Routing map[string]ModelRoute `yaml:"routing"`
	// PrefixRouting 包含基于前缀的路由规则
	PrefixRouting map[string]PrefixRoute `yaml:"prefixRouting"`
	// LoadBalancing 包含负载均衡配置
	LoadBalancing map[string]LoadBalanceConfig `yaml:"loadBalancing"`
}

// ModelRoute 定义模型请求应路由到的位置（精确匹配）。
type ModelRoute struct {
	Provider    string `yaml:"provider"`    // 供应商名称
	ActualModel string `yaml:"actualModel"` // 可选：要使用的实际模型名称
}

// PrefixRoute 定义基于前缀的路由（例如，"deepseek-" -> siliconflow）。
type PrefixRoute struct {
	Provider string `yaml:"provider"`
	Priority int    `yaml:"priority"` // 当多个前缀匹配时，优先级较高的胜出
}

// LoadBalanceConfig 定义模型模式的负载均衡。
type LoadBalanceConfig struct {
	Strategy  string              `yaml:"strategy"` // "round-robin" | "random" | "failover" | "weighted"
	Providers []LoadBalanceMember `yaml:"providers"`
}

// LoadBalanceMember 代表负载均衡组中的一个供应商。
type LoadBalanceMember struct {
	Name     string `yaml:"name"`     // 供应商名称
	Weight   int    `yaml:"weight"`   // 用于权重策略
	Priority int    `yaml:"priority"` // 用于故障转移策略（数值越小优先级越高）
}

// AuthConfig 包含身份验证设置。
type AuthConfig struct {
	Enabled   bool   `yaml:"enabled"`
	JWTSecret string `yaml:"jwtSecret"`
	// 注意：API Keys 现在在数据库中管理，不再从配置文件读取
}

// Load 从 YAML 文件读取配置。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 环境变量覆盖（优先级高于配置文件）
	overrideFromEnv(&cfg)

	// 设置默认值
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.HTTP.ReadTimeout == 0 {
		cfg.HTTP.ReadTimeout = 30 * time.Second
	}
	if cfg.HTTP.WriteTimeout == 0 {
		cfg.HTTP.WriteTimeout = 120 * time.Second
	}

	// 为供应商设置默认超时时间
	for i := range cfg.Providers {
		if cfg.Providers[i].Timeout == 0 {
			cfg.Providers[i].Timeout = 60 * time.Second
		}
	}

	// 如果映射为 nil，则初始化
	if cfg.Models.Routing == nil {
		cfg.Models.Routing = make(map[string]ModelRoute)
	}
	if cfg.Models.PrefixRouting == nil {
		cfg.Models.PrefixRouting = make(map[string]PrefixRoute)
	}
	if cfg.Models.LoadBalancing == nil {
		cfg.Models.LoadBalancing = make(map[string]LoadBalanceConfig)
	}

	// 为负载均衡成员设置默认权重
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

// overrideFromEnv 从环境变量覆盖配置。
// 支持的环境变量：
//   - DB_HOST: 数据库主机
//   - DB_PORT: 数据库端口
//   - DB_USER: 数据库用户
//   - DB_PASSWORD: 数据库密码
//   - DB_NAME: 数据库名称
//   - JWT_SECRET: JWT 密钥
func overrideFromEnv(cfg *Config) {
	// 数据库配置
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.MySQL.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		// 简单的端口号转换，生产环境应该有更完善的错误处理
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.MySQL.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.MySQL.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.MySQL.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.MySQL.Database = v
	}

	// JWT 密钥
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
}

// DefaultConfig 为开发环境返回默认配置。
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
