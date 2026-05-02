package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/require"
)

func TestNoKeyGlobalsInCommandAndSeeders(t *testing.T) {
	repoRoot := testkit.RepoRoot(t)

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
				require.NoError(t, err)
				if d.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}

				b, err := os.ReadFile(path)
				require.NoError(t, err)
				require.NotContainsf(t, string(b), r.needle, "disallowed reference found in %s", path)
				return nil
			})
			require.NoError(t, err)
		})
	}
}
