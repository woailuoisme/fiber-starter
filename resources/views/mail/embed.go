package mailviews

import "embed"

// FS exposes the embedded email template files.
//
//go:embed layouts/*.tmpl components/*.tmpl *.tmpl
var FS embed.FS
