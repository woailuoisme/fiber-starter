package services

import (
	"errors"
	"strings"
	"time"

	database "fiber-starter/database"

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

func userAllowedUpdates(updates map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	for k, v := range updates {
		field := strings.ToLower(strings.TrimSpace(k))
		switch field {
		case "name", "email", "avatar", "phone", "status", "email_verified_at":
			filtered[field] = v
		default:
		}
	}
	return filtered
}
