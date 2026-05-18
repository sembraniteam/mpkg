# mpkg

Generate static HTML for Go vanity import paths.

[![CI](https://github.com/sembraniteam/mpkg/actions/workflows/ci.yml/badge.svg)](https://github.com/sembraniteam/mpkg/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sembraniteam/mpkg.svg)](https://pkg.go.dev/go.sembraniteam.com/mpkg)
[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go)](https://go.dev)

**mpkg** reads a YAML config, generates a self-contained static site where every module page carries the correct `go-import` and `go-source` meta tags, and produces a searchable index listing all packages. Deploy the output directory to any static host — no server-side code required.

## Installation

```bash
go install go.sembraniteam.com/mpkg@latest
```

Or download a pre-built binary from [GitHub Releases](https://github.com/sembraniteam/mpkg/releases).

## Quick Start

**1. Create `mpkg.yaml`:**

```yaml
base_url: your-domain.com

site:
  title: "My Go Packages"
  tagline: "Custom vanity import paths"
  logo: "MyOrg"
  logo_font: "standard"

build:
  output_dir: build

modules:
  mylib:
    git_url: https://github.com/your-org/mylib
    branch: main
    description: "A short description of your package."
    tags: [library, go]
```

**2. Validate the config:**

```bash
mpkg validate
```

**3. Generate the site:**

```bash
mpkg generate
```

**4. Deploy `build/` to your static host** (Nginx, GitHub Pages, Caddy, etc.).

Users can now run:

```bash
go get your-domain.com/mylib@latest
```

## Configuration

All fields in `mpkg.yaml`:

| Field                        | Type     | Default    | Description                                                          |
|------------------------------|----------|------------|----------------------------------------------------------------------|
| `base_url`                   | string   | —          | Your domain, e.g. `example.com` (**required**)                       |
| `site.title`                 | string   | —          | HTML `<title>` of the index page                                     |
| `site.tagline`               | string   | —          | Subtitle shown below the logo                                        |
| `site.logo`                  | string   | `MPKG`     | Text rendered as ASCII art in the header                             |
| `site.logo_font`             | string   | `standard` | [go-figure](https://github.com/common-nighthawk/go-figure) font name |
| `build.output_dir`           | string   | `build`    | Directory where HTML files are written                               |
| `modules.<name>.git_url`     | string   | —          | Git repository URL (**required**)                                    |
| `modules.<name>.branch`      | string   | `main`     | Default branch                                                       |
| `modules.<name>.description` | string   | —          | Short description shown on the index page                            |
| `modules.<name>.tags`        | []string | —          | Tags used for search filtering                                       |

## Commands

| Command      | Description                                         |
|--------------|-----------------------------------------------------|
| `validate`   | Validate `mpkg.yaml` without generating any output  |
| `generate`   | Build static HTML into the output directory         |
| `serve`      | Preview the generated site with a local HTTP server |
| `watch`      | Serve and auto-regenerate when `mpkg.yaml` changes  |
| `version`    | Print the mpkg version                              |
| `completion` | Generate shell autocompletion script                |

### Global flag

| Flag           | Default     | Description         |
|----------------|-------------|---------------------|
| `-c, --config` | `mpkg.yaml` | Path to config file |

### `generate`

```bash
mpkg generate [flags]
```

| Flag           | Default         | Description                                   |
|----------------|-----------------|-----------------------------------------------|
| `--clean`      | `false`         | Remove the output directory before generating |
| `-o, --output` | _(from config)_ | Override the output directory                 |
| `-q, --quiet`  | `false`         | Suppress all output except errors             |

### `serve`

```bash
mpkg serve [flags]
```

| Flag         | Default | Description                                            |
|--------------|---------|--------------------------------------------------------|
| `-p, --port` | `8080`  | HTTP port to listen on                                 |
| `--open`     | `false` | Open the browser automatically after the server starts |

### `watch`

```bash
mpkg watch [flags]
```

| Flag         | Default | Description                                            |
|--------------|---------|--------------------------------------------------------|
| `-p, --port` | `8080`  | HTTP port to listen on                                 |
| `--open`     | `false` | Open the browser automatically after the server starts |

## Contributing

```bash
git clone https://github.com/sembraniteam/mpkg.git
cd mpkg
go test ./...   # run the full test suite
go build .      # build the binary
```

Open an issue before submitting a large change. PRs are welcome.
