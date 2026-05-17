package drivers

import (
	"database/sql"
	"net/url"
	"path/filepath"
	"strings"

	"fiber-starter/configs"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConnector struct{}

func (c *SQLiteConnector) DriverName() string {
	return "sqlite3"
}

func (c *SQLiteConnector) Connect(cfg configs.DBConnection) (*sql.DB, error) {
	path := strings.TrimSpace(cfg.Database)
	if path == "" {
		path = ":memory:"
	}

	if path != ":memory:" && !strings.HasPrefix(path, "file:") {
		if !filepath.IsAbs(path) {
			path = filepath.Clean(path)
		}
	}

	var dsn string
	if strings.HasPrefix(path, "file:") {
		dsn = path
	} else {
		u := url.URL{Scheme: "file", Path: path}
		q := u.Query()
		q.Set("_foreign_keys", "1")
		q.Set("_busy_timeout", "5000")
		q.Set("parseTime", "true")
		q.Set("_loc", "UTC")
		u.RawQuery = q.Encode()
		dsn = u.String()
	}

	return sql.Open(c.DriverName(), dsn)
}
