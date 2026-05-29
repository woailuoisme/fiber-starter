package medialibrary

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

type FileNamePolicy struct {
	DisallowedExtensions []string
}

func NewFileNamePolicy() FileNamePolicy {
	return FileNamePolicy{
		DisallowedExtensions: []string{
			"php", "php3", "php4", "php5", "php6", "php7", "php8",
			"phtml", "phtm", "phar", "cgi", "pl", "asp", "aspx", "jsp", "jspx",
			"htaccess", "htpasswd",
		},
	}
}

func (p FileNamePolicy) Sanitize(name string) (string, error) {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		switch r {
		case '/', '\\', '#':
			return '-'
		case ' ':
			return '_'
		default:
			return r
		}
	}, name)
	name = strings.Trim(name, ".")
	if name == "" {
		return "", fmt.Errorf("%w: empty file name", ErrFileNameNotAllowed)
	}

	parts := strings.Split(strings.ToLower(name), ".")
	if len(parts) > 1 {
		for _, ext := range parts[1:] {
			if p.isDisallowed(ext) {
				return "", fmt.Errorf("%w: extension %q is not allowed", ErrFileNameNotAllowed, ext)
			}
		}
	}

	return name, nil
}

func (p FileNamePolicy) isDisallowed(ext string) bool {
	for _, disallowed := range p.DisallowedExtensions {
		if ext == disallowed {
			return true
		}
	}
	return false
}
