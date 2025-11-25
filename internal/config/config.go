package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the complete application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Schemas  SchemasConfig  `mapstructure:"schemas"`
	Limits   LimitsConfig   `mapstructure:"limits"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Security SecurityConfig `mapstructure:"security"`
	Features FeaturesConfig `mapstructure:"features"`
	API      APIConfig      `mapstructure:"api"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedOrigins []string `mapstructure:"allowedOrigins"`
	AllowedMethods []string `mapstructure:"allowedMethods"`
}

// SchemasConfig holds schema loading configuration
type SchemasConfig struct {
	LoadFromFiles bool   `mapstructure:"loadFromFiles"`
	Directory     string `mapstructure:"directory"`
}

// LimitsConfig holds various limits
type LimitsConfig struct {
	MaxQueryLength     int             `mapstructure:"maxQueryLength"`
	MaxParameterCount  int             `mapstructure:"maxParameterCount"`
	MaxParseDepth      int             `mapstructure:"maxParseDepth"`
	MaxSchemaFields    int             `mapstructure:"maxSchemaFields"`
	MaxFieldNameLength int             `mapstructure:"maxFieldNameLength"`
	MaxSchemas         int             `mapstructure:"maxSchemas"`
	MaxRequestBodySize int64           `mapstructure:"maxRequestBodySize"`
	RequestTimeout     time.Duration   `mapstructure:"requestTimeout"`
	RateLimit          RateLimitConfig `mapstructure:"rateLimit"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
	RequestsPerHour   int  `mapstructure:"requestsPerHour"`
	Burst             int  `mapstructure:"burst"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"`
	MaxSize int  `mapstructure:"maxSize"`
	TTL     int  `mapstructure:"ttl"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	AllowedSpecialChars string     `mapstructure:"allowedSpecialChars"`
	BlockSqlKeywords    bool       `mapstructure:"blockSqlKeywords"`
	Auth                AuthConfig `mapstructure:"auth"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Type    string   `mapstructure:"type"`
	APIKeys []string `mapstructure:"apiKeys"`
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	QuerySuggestions bool   `mapstructure:"querySuggestions"`
	MaxQueryLength   int    `mapstructure:"maxQueryLength"`
	RequestIDHeader  string `mapstructure:"requestIdHeader"`
}

// APIConfig holds API configuration
type APIConfig struct {
	Versions map[string]APIVersionConfig `mapstructure:"versions"`
}

// APIVersionConfig holds configuration for a specific API version
type APIVersionConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	Deprecated bool `mapstructure:"deprecated"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure viper
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/rsearch/")
		v.AddConfigPath("$HOME/.rsearch/")
	}

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	// Environment variables
	v.SetEnvPrefix("RSEARCH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.readTimeout", "30s")
	v.SetDefault("server.writeTimeout", "30s")
	v.SetDefault("server.shutdownTimeout", "10s")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics defaults
	v.SetDefault("metrics.enabled", false)
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")

	// CORS defaults
	v.SetDefault("cors.enabled", false)
	v.SetDefault("cors.allowedOrigins", []string{"*"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "DELETE"})

	// Schemas defaults
	v.SetDefault("schemas.loadFromFiles", false)
	v.SetDefault("schemas.directory", "./schemas")

	// Limits defaults
	v.SetDefault("limits.maxQueryLength", 10000)
	v.SetDefault("limits.maxParameterCount", 100)
	v.SetDefault("limits.maxParseDepth", 50)
	v.SetDefault("limits.maxSchemaFields", 1000)
	v.SetDefault("limits.maxFieldNameLength", 255)
	v.SetDefault("limits.maxSchemas", 100)
	v.SetDefault("limits.maxRequestBodySize", 1048576)
	v.SetDefault("limits.requestTimeout", "30s")
	v.SetDefault("limits.rateLimit.enabled", false)
	v.SetDefault("limits.rateLimit.requestsPerMinute", 100)
	v.SetDefault("limits.rateLimit.requestsPerHour", 5000)
	v.SetDefault("limits.rateLimit.burst", 10)

	// Cache defaults
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.maxSize", 10000)
	v.SetDefault("cache.ttl", 3600)

	// Security defaults
	v.SetDefault("security.allowedSpecialChars", ".-_")
	v.SetDefault("security.blockSqlKeywords", true)
	v.SetDefault("security.auth.enabled", false)
	v.SetDefault("security.auth.type", "apikey")
	v.SetDefault("security.auth.apiKeys", []string{})

	// Features defaults
	v.SetDefault("features.querySuggestions", false)
	v.SetDefault("features.maxQueryLength", 1000)
	v.SetDefault("features.requestIdHeader", "X-Request-ID")

	// API defaults
	v.SetDefault("api.versions.v1.enabled", true)
	v.SetDefault("api.versions.v1.deprecated", false)
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Server validation
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Logging validation
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or console)", cfg.Logging.Format)
	}

	// Metrics validation
	if cfg.Metrics.Enabled {
		if cfg.Metrics.Port < 1 || cfg.Metrics.Port > 65535 {
			return fmt.Errorf("invalid metrics port: %d", cfg.Metrics.Port)
		}
		if cfg.Metrics.Path == "" {
			return fmt.Errorf("metrics path cannot be empty when metrics are enabled")
		}
	}

	// Limits validation
	if cfg.Limits.MaxQueryLength < 0 {
		return fmt.Errorf("maxQueryLength cannot be negative")
	}
	if cfg.Limits.MaxParameterCount < 1 {
		return fmt.Errorf("maxParameterCount must be at least 1")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetMetricsAddress returns the metrics server address
func (c *Config) GetMetricsAddress() string {
	return fmt.Sprintf("localhost:%d", c.Metrics.Port)
}
