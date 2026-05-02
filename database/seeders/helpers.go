package seeders

import (
	"database/sql"
	"errors"

	dbsqlc "fiber-starter/database/sqlc"
)

func withSeedQueries(db *sql.DB, dialect string, fn func(*dbsqlc.Queries) error) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return fn(dbsqlc.New(db, dialect))
}
