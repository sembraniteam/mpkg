package generator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/common-nighthawk/go-figure"
	"go.sembraniteam.com/mpkg/internal/config"
	"go.sembraniteam.com/mpkg/templates"
)

const perm = 0o755

type indexData struct {
	BaseURL   string
	Title     string
	Tagline   string
	Logo      string
	Modules   []moduleRow
	Generated string
}

type moduleRow struct {
	Name        string
	FullPath    string
	Description string
	Tags        []string
	DocURL      string
	GitURL      string
	Branch      string
}

type moduleData struct {
	FullPath    string
	GitURL      string
	Branch      string
	Description string
}

type notFoundData struct {
	BaseURL string
}

// Build renders all pages into cfg.Build.OutputDir.
func Build(cfg *config.Config) error {
	tmpl, err := template.ParseFS(templates.FS, "*.tmpl")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}

	out := cfg.Build.OutputDir
	if err := os.MkdirAll(out, perm); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if err := renderIndex(tmpl, cfg, out); err != nil {
		return err
	}

	for name, mod := range cfg.Modules {
		if err := renderModule(tmpl, cfg.BaseURL, name, mod, out); err != nil {
			return err
		}
	}

	return render404(tmpl, cfg, out)
}

func renderIndex(
	tmpl *template.Template,
	cfg *config.Config,
	out string,
) error {
	rows := make([]moduleRow, 0, len(cfg.Modules))
	names := make([]string, 0, len(cfg.Modules))
	for n := range cfg.Modules {
		names = append(names, n)
	}
	sort.Slice(
		names,
		func(i, j int) bool { return naturalLess(names[i], names[j]) },
	)

	for _, name := range names {
		mod := cfg.Modules[name]
		rows = append(rows, moduleRow{
			Name:        name,
			FullPath:    cfg.BaseURL + "/" + name,
			Description: mod.Description,
			Tags:        mod.Tags,
			DocURL:      "https://pkg.go.dev/" + cfg.BaseURL + "/" + name,
			GitURL:      mod.GitURL,
			Branch:      mod.Branch,
		})
	}

	data := indexData{
		BaseURL:   cfg.BaseURL,
		Title:     cfg.Site.Title,
		Tagline:   cfg.Site.Tagline,
		Logo:      RenderLogo(cfg.Site.Logo, cfg.Site.LogoFont),
		Modules:   rows,
		Generated: time.Now().Format(time.DateOnly),
	}

	return writeTemplate(
		tmpl,
		"index.tmpl",
		filepath.Join(out, "index.html"),
		data,
	)
}

func renderModule(
	tmpl *template.Template,
	baseURL, name string,
	mod config.Module,
	out string,
) error {
	data := moduleData{
		FullPath:    baseURL + "/" + name,
		GitURL:      mod.GitURL,
		Branch:      mod.Branch,
		Description: mod.Description,
	}
	dir := filepath.Join(out, filepath.FromSlash(name))
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("create module dir %q: %w", dir, err)
	}
	return writeTemplate(
		tmpl,
		"module.tmpl",
		filepath.Join(dir, "index.html"),
		data,
	)
}

func render404(tmpl *template.Template, cfg *config.Config, out string) error {
	return writeTemplate(
		tmpl,
		"404.tmpl",
		filepath.Join(out, "404.html"),
		notFoundData{
			BaseURL: cfg.BaseURL,
		},
	)
}

func writeTemplate(tmpl *template.Template, name, dst string, data any) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer func() { _ = f.Close() }()
	if err := tmpl.ExecuteTemplate(f, name, data); err != nil {
		return fmt.Errorf("render %q: %w", name, err)
	}
	return nil
}

// RenderLogo converts plain text to ASCII art using go-figure.
// Falls back to "standard" font if the given font is not found.
func RenderLogo(text, font string) (art string) {
	if text == "" {
		text = "MPKG"
	}
	if font == "" {
		font = "standard"
	}
	defer func() {
		if r := recover(); r != nil {
			fig := figure.NewFigure(text, "standard", true)
			art = fig.String()
		}
	}()
	fig := figure.NewFigure(text, font, true)
	return fig.String()
}
