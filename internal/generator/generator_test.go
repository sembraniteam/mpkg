package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.sembraniteam.com/mpkg/internal/config"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		wantFiles []string
		wantIn    map[string]string
	}{
		{
			name: "generates index and module pages",
			cfg: &config.Config{
				BaseURL: "example.com",
				Site: config.SiteConfig{
					Title:    "Test",
					Tagline:  "Tagline",
					Logo:     "Test",
					LogoFont: "standard",
				},
				Build: config.BuildConfig{OutputDir: ""},
				Modules: map[string]config.Module{
					"router": {
						GitURL:      "https://github.com/test/router",
						Branch:      "main",
						Description: "HTTP router",
						Tags:        []string{"http"},
					},
				},
			},
			wantFiles: []string{"index.html", "router/index.html", "404.html"},
			wantIn: map[string]string{
				"index.html":        "example.com",
				"router/index.html": `go-import`,
				"404.html":          "Package not found",
			},
		},
		{
			name: "empty modules renders empty state",
			cfg: &config.Config{
				BaseURL: "example.com",
				Site:    config.SiteConfig{Logo: "Test", LogoFont: "standard"},
				Build:   config.BuildConfig{},
				Modules: map[string]config.Module{},
			},
			wantFiles: []string{"index.html", "404.html"},
			wantIn: map[string]string{
				"index.html": "No packages yet",
			},
		},
		{
			name: "nested module path creates nested dir",
			cfg: &config.Config{
				BaseURL: "example.com",
				Site:    config.SiteConfig{Logo: "T", LogoFont: "standard"},
				Build:   config.BuildConfig{},
				Modules: map[string]config.Module{
					"pkg/v2": {
						GitURL: "https://github.com/test/pkg",
						Branch: "main",
					},
				},
			},
			wantFiles: []string{"pkg/v2/index.html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.cfg.Build.OutputDir = dir

			err := Build(tt.cfg)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			for _, f := range tt.wantFiles {
				path := filepath.Join(dir, f)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("expected file %q not found: %v", f, err)
				}
			}

			for f, want := range tt.wantIn {
				path := filepath.Join(dir, f)
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %q: %v", f, err)
				}
				if !strings.Contains(string(content), want) {
					t.Errorf("file %q does not contain %q", f, want)
				}
			}
		})
	}
}

func TestRenderLogoFallback(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		font     string
		wantText string
	}{
		{name: "valid font", text: "Hi", font: "standard", wantText: "Hi"},
		{
			name:     "empty text uses default",
			text:     "",
			font:     "standard",
			wantText: "Hi",
		},
		{
			name:     "invalid font falls back",
			text:     "Hi",
			font:     "not_a_real_font_xyz",
			wantText: "Hi",
		},
		{
			name:     "empty font uses standard",
			text:     "Hi",
			font:     "",
			wantText: "Hi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderLogo(tt.text, tt.font)
			if got == "" {
				t.Error("RenderLogo() returned empty string")
			}
		})
	}
}

func TestBuildOutputDirNotCreatable(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o444); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BaseURL: "example.com",
		Build: config.BuildConfig{
			OutputDir: filepath.Join(blocker, "subdir"),
		},
		Modules: map[string]config.Module{},
	}
	err := Build(cfg)
	if err == nil {
		t.Error("expected error when output dir is not creatable, got nil")
	}
}

func TestBuildWriteTemplateError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	cfg := &config.Config{
		BaseURL: "example.com",
		Site:    config.SiteConfig{Logo: "T", LogoFont: "standard"},
		Build:   config.BuildConfig{OutputDir: dir},
		Modules: map[string]config.Module{},
	}
	err := Build(cfg)
	if err == nil {
		t.Error("expected error writing templates to read-only dir, got nil")
	}
}

func TestRenderIndexSortOrder(t *testing.T) {
	// Build 22 versioned modules: mpkg/v1 … mpkg/v22
	// Lexicographic sort would produce: v1, v10, v11, …, v19, v2, v20, v21,
	// v22, v3 … v9
	// Natural sort must produce:        v1, v2, v3, …, v9, v10, v11, …, v22
	modules := map[string]config.Module{}
	for i := 1; i <= 22; i++ {
		name := fmt.Sprintf("mpkg/v%d", i)
		modules[name] = config.Module{
			GitURL: "https://github.com/test/mpkg",
			Branch: "main",
		}
	}

	cfg := &config.Config{
		BaseURL: "example.com",
		Site:    config.SiteConfig{Logo: "T", LogoFont: "standard"},
		Build:   config.BuildConfig{OutputDir: t.TempDir()},
		Modules: modules,
	}

	if err := Build(cfg); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	content, err := os.ReadFile(
		filepath.Join(cfg.Build.OutputDir, "index.html"),
	)
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	html := string(content)

	// Check that each version appears before the next in the rendered HTML.
	// Natural order: v1 < v2 < … < v9 < v10 < v11 < … < v22
	order := make([]string, 22)
	for i := range order {
		order[i] = fmt.Sprintf("mpkg/v%d", i+1)
	}

	prev := -1
	for _, name := range order {
		pos := strings.Index(html, name)
		if pos == -1 {
			t.Errorf("module %q not found in index.html", name)
			continue
		}
		if pos <= prev {
			t.Errorf(
				"module %q appears at position %d, expected after position %d (wrong sort order)",
				name,
				pos,
				prev,
			)
		}
		prev = pos
	}
}

func TestBuildMetaTags(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		BaseURL: "example.com",
		Site: config.SiteConfig{
			Title:    "My Packages",
			Tagline:  "Go vanity imports",
			Logo:     "Test",
			LogoFont: "standard",
		},
		Build: config.BuildConfig{OutputDir: dir},
		Modules: map[string]config.Module{
			"router": {
				GitURL:      "https://github.com/test/router",
				Branch:      "main",
				Description: "HTTP router",
				Tags:        []string{"http"},
			},
		},
	}

	if err := Build(cfg); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	indexHTML, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	for _, want := range []string{
		`<meta name="description" content="Go vanity imports">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:title" content="My Packages">`,
		`<meta property="og:description" content="Go vanity imports">`,
		`<meta property="og:url" content="https://example.com">`,
	} {
		if !strings.Contains(string(indexHTML), want) {
			t.Errorf("index.html missing %q", want)
		}
	}

	moduleHTML, err := os.ReadFile(filepath.Join(dir, "router/index.html"))
	if err != nil {
		t.Fatalf("read router/index.html: %v", err)
	}
	for _, want := range []string{
		`<meta name="description" content="HTTP router">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:title" content="example.com/router">`,
		`<meta property="og:description" content="HTTP router">`,
		`<meta property="og:url" content="https://example.com/router">`,
		`<meta name="robots" content="noindex, follow">`,
	} {
		if !strings.Contains(string(moduleHTML), want) {
			t.Errorf("router/index.html missing %q", want)
		}
	}
}
