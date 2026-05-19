package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid full config",
			yaml: `
base_url: geekswamp.dev
site:
  title: "Test Packages"
  tagline: "My packages"
  logo: "Test"
  logo_font: "slant"
build:
  output_dir: dist
modules:
  router:
    git_url: https://github.com/test/router
    branch: main
    description: "HTTP router"
    tags: [http, web]
`,
			want: &Config{
				BaseURL: "geekswamp.dev",
				Site: SiteConfig{
					Title:    "Test Packages",
					Tagline:  "My packages",
					Logo:     "Test",
					LogoFont: "slant",
				},
				Build: BuildConfig{OutputDir: "dist"},
				Modules: map[string]Module{
					"router": {
						GitURL:      "https://github.com/test/router",
						Branch:      "main",
						Description: "HTTP router",
						Tags:        []string{"http", "web"},
					},
				},
			},
		},
		{
			name:    "missing base_url",
			yaml:    "site:\n  title: Test\nmodules:\n  r:\n    git_url: https://g.com/r\n",
			wantErr: true,
		},
		{
			name:    "missing git_url",
			yaml:    "base_url: example.com\nmodules:\n  router:\n    description: test\n",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			yaml:    "base_url: [invalid",
			wantErr: true,
		},
		{
			name: "empty modules is valid",
			yaml: "base_url: example.com\nmodules: {}\n",
			want: &Config{
				BaseURL: "example.com",
				Site:    SiteConfig{Logo: "MPKG", LogoFont: "standard"},
				Build:   BuildConfig{OutputDir: "build"},
				Modules: map[string]Module{},
			},
		},
		{
			name: "base_url with https scheme is normalized",
			yaml: "base_url: https://example.com\nmodules: {}\n",
			want: &Config{
				BaseURL: "example.com",
				Site:    SiteConfig{Logo: "MPKG", LogoFont: "standard"},
				Build:   BuildConfig{OutputDir: "build"},
				Modules: map[string]Module{},
			},
		},
		{
			name: "base_url with http scheme and trailing slash is normalized",
			yaml: "base_url: http://example.com/\nmodules: {}\n",
			want: &Config{
				BaseURL: "example.com",
				Site:    SiteConfig{Logo: "MPKG", LogoFont: "standard"},
				Build:   BuildConfig{OutputDir: "build"},
				Modules: map[string]Module{},
			},
		},
		{
			name: "defaults applied: branch, output_dir, logo",
			yaml: "base_url: example.com\nmodules:\n  pkg:\n    git_url: https://g.com/pkg\n",
			want: &Config{
				BaseURL: "example.com",
				Site:    SiteConfig{Logo: "MPKG", LogoFont: "standard"},
				Build:   BuildConfig{OutputDir: "build"},
				Modules: map[string]Module{
					"pkg": {
						GitURL: "https://g.com/pkg",
						Branch: "main",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFromReader(strings.NewReader(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"LoadFromReader() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if got.BaseURL != tt.want.BaseURL {
				t.Errorf("BaseURL = %q, want %q", got.BaseURL, tt.want.BaseURL)
			}
			if got.Site != tt.want.Site {
				t.Errorf("Site = %+v, want %+v", got.Site, tt.want.Site)
			}
			if got.Build != tt.want.Build {
				t.Errorf("Build = %+v, want %+v", got.Build, tt.want.Build)
			}
			for k, want := range tt.want.Modules {
				gotMod, ok := got.Modules[k]
				if !ok {
					t.Errorf("missing module %q", k)
					continue
				}
				if gotMod.GitURL != want.GitURL ||
					gotMod.Branch != want.Branch ||
					gotMod.Description != want.Description {
					t.Errorf("module %q = %+v, want %+v", k, gotMod, want)
				}
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid file",
			content: "base_url: example.com\nmodules: {}\n",
		},
		{
			name:    "nonexistent file",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != "" {
				f, err := os.CreateTemp(t.TempDir(), "mpkg*.yaml")
				if err != nil {
					t.Fatalf("create temp file: %v", err)
				}
				_, _ = f.WriteString(tt.content)
				_ = f.Close()
				path = f.Name()
			} else {
				path = "/nonexistent/path/mpkg.yaml"
			}

			_, err := Load(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "with module",
			err: &ValidationError{
				Module: "router",
				Field:  "git_url",
				Msg:    "is required",
			},
			want: `module "router": git_url is required`,
		},
		{
			name: "without module",
			err:  &ValidationError{Field: "base_url", Msg: "is required"},
			want: "base_url: is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
