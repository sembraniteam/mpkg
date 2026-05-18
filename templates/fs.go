package templates

import "embed"

// FS holds all embedded Go template files.
//
//go:embed *.tmpl
var FS embed.FS
