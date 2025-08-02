package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Auth     AuthConfig     `yaml:"auth"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port                  int    `yaml:"port"`
	Host                  string `yaml:"host"`
	ReadTimeoutSeconds    int    `yaml:"readTimeoutSeconds"`
	WriteTimeoutSeconds   int    `yaml:"writeTimeoutSeconds"`
	IdleTimeoutSeconds    int    `yaml:"idleTimeoutSeconds"`
	ShutdownTimeoutSeconds int    `yaml:"shutdownTimeoutSeconds"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslMode"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	File   string `yaml:"file"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret            string `yaml:"jwtSecret"`
	JWTExpirationMinutes int    `yaml:"jwtExpirationMinutes"`
}

// Load reads the configuration from a file
func Load(path string) (*Config, error) {
	// Set default configuration
	config := &Config{
		Server: ServerConfig{
			Port:                  8080,
			Host:                  "localhost",
			ReadTimeoutSeconds:    30,
			WriteTimeoutSeconds:   30,
			IdleTimeoutSeconds:    60,
			ShutdownTimeoutSeconds: 30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// Read configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML configuration
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}
