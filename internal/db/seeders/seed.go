// Package seeders 处理数据库种子数据的生成和填充
package seeders

import (
	"database/sql"
	"errors"
	"log"

	"gorm.io/gorm"
)

// RunAllSeeders 运行所有种子数据
func RunAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to execute seed data...")

	if err := withSeedDB(db, dialect, SeedUsers); err != nil {
		log.Printf("User seed data creation failed: %v", err)
		return err
	}
	log.Println("User seed data creation completed")

	// Add other table seed data here
	// if err := SeedPosts(db); err != nil {
	//     log.Printf("Post seed data creation failed: %v", err)
	//     return err
	// }
	// log.Println("Post seed data creation completed")

	log.Println("All seed data creation completed")
	return nil
}

// RunRandomSeeders Run random seed data generation
func RunRandomSeeders(db *sql.DB, dialect string, userCount int) error {
	log.Printf("Starting to generate %d random user data...", userCount)

	if err := withSeedDB(db, dialect, func(sqlDB *sql.DB, driver string) error {
		return GenerateTestUsers(sqlDB, driver, userCount)
	}); err != nil {
		log.Printf("Random user data generation failed: %v", err)
		return err
	}
	log.Printf("Successfully generated %d random user data", userCount)

	return nil
}

// ClearAllSeeders Clear all seed data
func ClearAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to clear seed data...")

	if err := withSeedDB(db, dialect, ClearUsers); err != nil {
		log.Printf("User seed data clearing failed: %v", err)
		return err
	}
	log.Println("User seed data clearing completed")

	// Add other table seed data clearing here
	// if err := ClearPosts(db); err != nil {
	//     log.Printf("Post seed data clearing failed: %v", err)
	//     return err
	// }
	// log.Println("Post seed data clearing completed")

	log.Println("All seed data clearing completed")
	return nil
}

// RefreshAllSeeders Refresh all seed data (clear first then create)
func RefreshAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to refresh seed data...")

	if err := ClearAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data clearing failed: %v", err)
		return err
	}

	if err := RunAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data creation failed: %v", err)
		return err
	}

	log.Println("Seed data refresh completed")
	return nil
}

// SetupDatabase Setup database (migration + seed data)
func SetupDatabase(db *sql.DB, dialect string) error {
	log.Println("Starting to setup database...")

	if err := RunAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data creation failed: %v", err)
		return err
	}

	log.Println("Database setup completed")
	return nil
}

func withSeedDB(db *sql.DB, dialect string, fn func(*sql.DB, string) error) error {
	if db == nil {
		return errors.New("db is nil")
	}

	return fn(db, dialect)
}

func withSeedGormDB(db *sql.DB, dialect string, fn func(*gorm.DB) error) error {
	gdb, err := seedGormDB(db, dialect)
	if err != nil {
		return err
	}

	return fn(gdb)
}
