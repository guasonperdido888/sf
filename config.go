package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Group represents a target chat with an optional human-readable name.
type Group struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Config holds all bot configuration loaded from JSON.
type Config struct {
	BotToken  string  `json:"bot_token"`
	Groups    []Group `json:"groups"`
	Message   string  `json:"message"`
	ParseMode string  `json:"parse_mode"`
	MediaFile string  `json:"media_file,omitempty"`
	MediaType string  `json:"media_type,omitempty"`
}

// LoadConfig reads and parses the JSON configuration file.
// Returns an error if the file is missing, malformed, or invalid.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("bot_token is required")
	}
	if c.Message == "" {
		return fmt.Errorf("message is required")
	}
	if c.ParseMode == "" {
		c.ParseMode = "MarkdownV2"
	}
	return nil
}
