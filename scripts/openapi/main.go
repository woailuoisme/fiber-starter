package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	root, err := findRepoRoot()
	if err != nil {
		fatal(err)
	}

	sourcePath := filepath.Join(root, "docs", "openapi.yaml")
	targetPath := filepath.Join(root, "docs", "openapi.json")

	specBytes, err := os.ReadFile(sourcePath) // #nosec G304 -- sourcePath is derived from the repository root and a fixed file name.
	if err != nil {
		fatal(fmt.Errorf("read openapi source: %w", err))
	}

	var spec map[string]any
	if err := yaml.Unmarshal(specBytes, &spec); err != nil {
		fatal(fmt.Errorf("parse openapi yaml: %w", err))
	}

	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fatal(fmt.Errorf("render openapi json: %w", err))
	}
	jsonBytes = append(jsonBytes, '\n')

	if err := os.WriteFile(targetPath, jsonBytes, 0o644); err != nil { // #nosec G306 -- generated documentation is intentionally repository-readable.
		fatal(fmt.Errorf("write openapi json: %w", err))
	}
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, ".git")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("repository root not found")
		}
		wd = parent
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
