package drivers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"lfiber/configs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConnector struct{}

func (c *PostgresConnector) DriverName() string {
	return "pgx"
}

func (c *PostgresConnector) Connect(cfg configs.DBConnection) (*sql.DB, error) {
	host := strings.TrimSpace(cfg.Host)
	port := strings.TrimSpace(cfg.Port)
	user := strings.TrimSpace(cfg.Username)
	pass := cfg.Password
	dbname := strings.TrimSpace(cfg.Database)

	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbname == "" {
		return nil, errors.New("postgres database name is empty")
	}

	sslmode := strings.TrimSpace(cfg.SSLMode)
	if sslmode == "" {
		sslmode = "disable"
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbname,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()

	return sql.Open(c.DriverName(), u.String())
}
