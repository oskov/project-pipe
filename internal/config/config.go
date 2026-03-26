package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	LLM    LLMConfig    `yaml:"llm"`
	Git    GitConfig    `yaml:"git"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"` // debug, info, warn, error
	JSONLog  bool   `yaml:"json_log"`
}

type DBConfig struct {
	Driver string `yaml:"driver"` // sqlite, postgres
	DSN    string `yaml:"dsn"`
}

type LLMConfig struct {
	Provider string `yaml:"provider"` // openai, anthropic, ollama
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"` // for ollama or custom endpoints
}

// GitConfig holds settings for git operations and GitHub access.
type GitConfig struct {
	ProjectsDir string `yaml:"projects_dir"` // local root where repos are cloned
	GithubToken string `yaml:"github_token"` // personal access token for private repos
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// expand environment variables
	data = []byte(os.ExpandEnv(string(data)))

	cfg := &Config{
		Server: ServerConfig{Port: 8080, LogLevel: "info", JSONLog: false},
		DB:     DBConfig{Driver: "sqlite", DSN: "project-pipe.db"},
		LLM:    LLMConfig{Provider: "openai", Model: "gpt-4o-mini"},
		Git:    GitConfig{ProjectsDir: "./projects"},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
