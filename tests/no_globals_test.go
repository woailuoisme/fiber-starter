package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoKeyGlobalsInCommandAndSeeders(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join(".."))

	type rule struct {
		name    string
		rootDir string
		needle  string
	}

	rules := []rule{
		{name: "cli should not use config.GlobalConfig", rootDir: filepath.Join(repoRoot, "app", "Console", "Commands"), needle: "config.GlobalConfig"},
		{name: "seeders should not use config.GlobalConfig", rootDir: filepath.Join(repoRoot, "database", "seeders"), needle: "config.GlobalConfig"},
		{name: "seeders should not use database.GetDB", rootDir: filepath.Join(repoRoot, "database", "seeders"), needle: "database.GetDB("},
		{name: "cli should not use database.GetDB", rootDir: filepath.Join(repoRoot, "app", "Console", "Commands"), needle: "database.GetDB("},
		{name: "cli should not use database.DB", rootDir: filepath.Join(repoRoot, "app", "Console", "Commands"), needle: "database.DB"},
		{name: "seeders should not use database.DB", rootDir: filepath.Join(repoRoot, "database", "seeders"), needle: "database.DB"},
	}

	for _, r := range rules {
		t.Run(r.name, func(t *testing.T) {
			err := filepath.WalkDir(r.rootDir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if !strings.HasSuffix(path, ".go") {
					return nil
				}

				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				if strings.Contains(string(b), r.needle) {
					t.Fatalf("Disallowed reference found: %s in %s", r.needle, path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}
		})
	}
}
