package seeders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	models "fiber-starter/app/Models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func withSeedGormDB(db *sql.DB, dialect string, fn func(*gorm.DB) error) error {
	gdb, err := seedGormDB(db, dialect)
	if err != nil {
		return err
	}

	return fn(gdb)
}

func seedCountUsers(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func seedUserExistsByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func seedGormDB(db *sql.DB, dialect string) (*gorm.DB, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	switch strings.ToLower(strings.TrimSpace(dialect)) {
	case "psql", "postgres", "postgresql":
		return gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Dialector{Conn: db}, &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}
