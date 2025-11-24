package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Load config without file (should use defaults)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default values
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", cfg.Logging.Format)
	}

	if cfg.Metrics.Enabled != false {
		t.Errorf("Expected metrics disabled by default, got %v", cfg.Metrics.Enabled)
	}

	if cfg.Limits.MaxQueryLength != 10000 {
		t.Errorf("Expected default max query length 10000, got %d", cfg.Limits.MaxQueryLength)
	}
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("RSEARCH_SERVER_PORT", "9000")
	os.Setenv("RSEARCH_LOGGING_LEVEL", "debug")
	os.Setenv("RSEARCH_METRICS_ENABLED", "true")
	defer func() {
		os.Unsetenv("RSEARCH_SERVER_PORT")
		os.Unsetenv("RSEARCH_LOGGING_LEVEL")
		os.Unsetenv("RSEARCH_METRICS_ENABLED")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000 from env var, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug' from env var, got '%s'", cfg.Logging.Level)
	}

	if cfg.Metrics.Enabled != true {
		t.Errorf("Expected metrics enabled from env var, got %v", cfg.Metrics.Enabled)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyConfig func(*Config)
		expectError bool
	}{
		{
			name: "valid config",
			modifyConfig: func(c *Config) {
				// No modifications
			},
			expectError: false,
		},
		{
			name: "invalid port too low",
			modifyConfig: func(c *Config) {
				c.Server.Port = 0
			},
			expectError: true,
		},
		{
			name: "invalid port too high",
			modifyConfig: func(c *Config) {
				c.Server.Port = 70000
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			modifyConfig: func(c *Config) {
				c.Logging.Level = "invalid"
			},
			expectError: true,
		},
		{
			name: "invalid log format",
			modifyConfig: func(c *Config) {
				c.Logging.Format = "invalid"
			},
			expectError: true,
		},
		{
			name: "negative max query length",
			modifyConfig: func(c *Config) {
				c.Limits.MaxQueryLength = -1
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Host:            "localhost",
					Port:            8080,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					ShutdownTimeout: 10 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
					Output: "stdout",
				},
				Metrics: MetricsConfig{
					Enabled: true,
					Port:    9090,
					Path:    "/metrics",
				},
				Limits: LimitsConfig{
					MaxQueryLength:    10000,
					MaxParameterCount: 1,
				},
			}

			tt.modifyConfig(cfg)
			err := validate(cfg)

			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestGetAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}

	addr := cfg.GetAddress()
	expected := "0.0.0.0:8080"
	if addr != expected {
		t.Errorf("Expected address '%s', got '%s'", expected, addr)
	}
}

func TestGetMetricsAddress(t *testing.T) {
	cfg := &Config{
		Metrics: MetricsConfig{
			Port: 9090,
		},
	}

	addr := cfg.GetMetricsAddress()
	expected := "localhost:9090"
	if addr != expected {
		t.Errorf("Expected metrics address '%s', got '%s'", expected, addr)
	}
}
