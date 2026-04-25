package services

import (
	"errors"
	"time"

	database "fiber-starter/internal/db"

	"gorm.io/gorm"
)

func withGormDB(conn *database.Connection, fn func(*gorm.DB) error) error {
	if conn == nil {
		return errors.New("database connection is nil")
	}

	db, err := conn.GetGormDB()
	if err != nil {
		return err
	}

	return fn(db)
}

func utcNow() time.Time {
	return time.Now().UTC()
}

func normalizePagination(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return page, limit, (page - 1) * limit
}
