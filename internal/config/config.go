package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Anthropic  AnthropicConfig  `yaml:"anthropic"`
	Jira       JiraConfig       `yaml:"jira"`
	Processing ProcessingConfig `yaml:"processing"`
}

// AnthropicConfig represents Anthropic API configuration
type AnthropicConfig struct {
	APIKey            string `yaml:"api_key"`
	Model             string `yaml:"model"`
	TimeoutSeconds    int    `yaml:"timeout_seconds"`
	MaxTokens         int    `yaml:"max_tokens"`
	ChunkSizeChars    int    `yaml:"chunk_size_chars"`
	RetryCount        int    `yaml:"retry_count"`
	RetryDelaySeconds int    `yaml:"retry_delay_seconds"`
}

// JiraConfig represents JIRA API configuration
type JiraConfig struct {
	BaseURL    string `yaml:"base_url"`
	Username   string `yaml:"username"`
	APIToken   string `yaml:"api_token"`
	ProjectKey string `yaml:"project_key"`
	Timeout    int    `yaml:"timeout_seconds"`
}

// ProcessingConfig represents processing configuration
type ProcessingConfig struct {
	Mode             string `yaml:"mode"`
	OutputDir        string `yaml:"output_dir"`
	SaveIntermediate bool   `yaml:"save_intermediate"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Anthropic.APIKey == "" {
		return fmt.Errorf("anthropic API key is required")
	}

	if c.Jira.BaseURL == "" {
		return fmt.Errorf("JIRA base URL is required")
	}

	if c.Jira.Username == "" {
		return fmt.Errorf("JIRA username is required")
	}

	if c.Jira.APIToken == "" {
		return fmt.Errorf("JIRA API token is required")
	}

	if c.Jira.ProjectKey == "" {
		return fmt.Errorf("JIRA project key is required")
	}

	return nil
}
