package seeders

import (
	"database/sql"
	"log"
)

// DatabaseSeeder 负责协调所有数据库种子任务
type DatabaseSeeder struct{}

// NewDatabaseSeeder 创建数据库种子协调器
func NewDatabaseSeeder() *DatabaseSeeder {
	return &DatabaseSeeder{}
}

// RunAllSeeders 执行所有种子任务
func (s *DatabaseSeeder) RunAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to execute seed data...")

	if err := WithUserSeeder().SeedUsers(db, dialect); err != nil {
		log.Printf("User seed data creation failed: %v", err)
		return err
	}
	log.Println("User seed data creation completed")

	log.Println("All seed data creation completed")
	return nil
}

// RunRandomSeeders 生成随机种子数据
func (s *DatabaseSeeder) RunRandomSeeders(db *sql.DB, dialect string, userCount int) error {
	log.Printf("Starting to generate %d random user data...", userCount)

	if err := WithUserSeeder().SeedRandomUsers(db, dialect, userCount); err != nil {
		log.Printf("Random user data generation failed: %v", err)
		return err
	}
	log.Printf("Successfully generated %d random user data", userCount)

	return nil
}

// ClearAllSeeders 清空所有种子数据
func (s *DatabaseSeeder) ClearAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to clear seed data...")

	if err := WithUserSeeder().ClearUsers(db, dialect); err != nil {
		log.Printf("User seed data clearing failed: %v", err)
		return err
	}
	log.Println("User seed data clearing completed")

	log.Println("All seed data clearing completed")
	return nil
}

// RefreshAllSeeders 刷新所有种子数据
func (s *DatabaseSeeder) RefreshAllSeeders(db *sql.DB, dialect string) error {
	log.Println("Starting to refresh seed data...")

	if err := s.ClearAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data clearing failed: %v", err)
		return err
	}

	if err := s.RunAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data creation failed: %v", err)
		return err
	}

	log.Println("Seed data refresh completed")
	return nil
}

// SetupDatabase 初始化数据库种子数据
func (s *DatabaseSeeder) SetupDatabase(db *sql.DB, dialect string) error {
	log.Println("Starting to setup database...")

	if err := s.RunAllSeeders(db, dialect); err != nil {
		log.Printf("Seed data creation failed: %v", err)
		return err
	}

	log.Println("Database setup completed")
	return nil
}

// Package-level compatibility helpers
func RunAllSeeders(db *sql.DB, dialect string) error {
	return NewDatabaseSeeder().RunAllSeeders(db, dialect)
}

func RunRandomSeeders(db *sql.DB, dialect string, userCount int) error {
	return NewDatabaseSeeder().RunRandomSeeders(db, dialect, userCount)
}

func ClearAllSeeders(db *sql.DB, dialect string) error {
	return NewDatabaseSeeder().ClearAllSeeders(db, dialect)
}

func RefreshAllSeeders(db *sql.DB, dialect string) error {
	return NewDatabaseSeeder().RefreshAllSeeders(db, dialect)
}

func SetupDatabase(db *sql.DB, dialect string) error {
	return NewDatabaseSeeder().SetupDatabase(db, dialect)
}
