package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Config is the top-level configuration parsed from mpkg.yaml.
type Config struct {
	BaseURL string            `yaml:"base_url"`
	Site    SiteConfig        `yaml:"site"`
	Build   BuildConfig       `yaml:"build"`
	Modules map[string]Module `yaml:"modules"`
}

// SiteConfig holds branding settings.
type SiteConfig struct {
	Title    string `yaml:"title"`
	Tagline  string `yaml:"tagline"`
	Logo     string `yaml:"logo"`
	LogoFont string `yaml:"logo_font"`
}

// BuildConfig holds output settings.
type BuildConfig struct {
	OutputDir string `yaml:"output_dir"`
}

// Module describes a single Go package entry.
type Module struct {
	GitURL      string   `yaml:"git_url"`
	Branch      string   `yaml:"branch"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// Load reads and parses a config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = f.Close() }()
	return LoadFromReader(f)
}

// LoadFromReader parses config from an io.Reader (useful for testing).
func LoadFromReader(r io.Reader) (*Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func normalizeBaseURL(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimRight(u, "/")
	return u
}

func applyDefaults(cfg *Config) {
	cfg.BaseURL = normalizeBaseURL(cfg.BaseURL)
	if cfg.Site.Logo == "" {
		cfg.Site.Logo = "MPKG"
	}
	if cfg.Site.LogoFont == "" {
		cfg.Site.LogoFont = "standard"
	}
	if cfg.Build.OutputDir == "" {
		cfg.Build.OutputDir = "build"
	}
	if cfg.Modules == nil {
		cfg.Modules = map[string]Module{}
	}
	for name, mod := range cfg.Modules {
		if mod.Branch == "" {
			mod.Branch = "main"
		}
		cfg.Modules[name] = mod
	}
}

func validate(cfg *Config) error {
	if cfg.BaseURL == "" {
		return &ValidationError{Field: "base_url", Msg: "is required"}
	}
	for name, mod := range cfg.Modules {
		if mod.GitURL == "" {
			return &ValidationError{
				Module: name,
				Field:  "git_url",
				Msg:    "is required",
			}
		}
	}
	return nil
}
