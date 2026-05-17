package Drivers

import (
	"context"
	"database/sql"
	"fmt"

	exceptions "fiber-starter/app/Exceptions"
	models "fiber-starter/app/Models"
	database "fiber-starter/app/Providers/Database/Contracts"
	hashContracts "fiber-starter/app/Providers/Hash/Contracts"
	repositories "fiber-starter/app/Repositories"
)

// DatabaseUserProvider implements the UserProvider interface using the application database
type DatabaseUserProvider struct {
	db     database.Connection
	table  string
	hasher hashContracts.Hasher
}

// NewDatabaseUserProvider creates a new database user provider instance
func NewDatabaseUserProvider(db database.Connection, table string, hasher hashContracts.Hasher) *DatabaseUserProvider {
	return &DatabaseUserProvider{db: db, table: table, hasher: hasher}
}

// RetrieveById retrieves a user by their unique identifier
func (p *DatabaseUserProvider) RetrieveById(id int64) (*models.User, error) {
	if p == nil || p.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", fmt.Errorf("database connection not initialized"))
	}

	if p.table == "users" {
		bunDB, err := p.db.BunDB()
		if err != nil {
			return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
		}
		userRepo := repositories.NewUserRepository(bunDB)
		user, err := userRepo.GetByID(context.Background(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return user, nil
	}

	// Generic raw SQL for other tables
	db, err := p.db.GetDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}

	// nolint:gosec // table name is from trusted config
	query := fmt.Sprintf("SELECT id, name, email, password FROM %s WHERE id = $1 AND deleted_at IS NULL LIMIT 1", p.table)
	row := db.QueryRowContext(context.Background(), query, id)

	var user models.User
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// RetrieveByCredentials retrieves a user by the given credentials (e.g. email)
func (p *DatabaseUserProvider) RetrieveByCredentials(credentials map[string]string) (*models.User, error) {
	if p == nil || p.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", fmt.Errorf("database connection not initialized"))
	}

	email, ok := credentials["email"]
	if !ok {
		return nil, nil
	}

	if p.table == "users" {
		bunDB, err := p.db.BunDB()
		if err != nil {
			return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
		}
		userRepo := repositories.NewUserRepository(bunDB)
		user, err := userRepo.GetByEmail(context.Background(), email)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return user, nil
	}

	// Generic raw SQL for other tables
	db, err := p.db.GetDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}

	// nolint:gosec // table name is from trusted config
	query := fmt.Sprintf("SELECT id, name, email, password FROM %s WHERE email = $1 AND deleted_at IS NULL LIMIT 1", p.table)
	row := db.QueryRowContext(context.Background(), query, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (p *DatabaseUserProvider) ValidateCredentials(user *models.User, credentials map[string]string) bool {
	password, ok := credentials["password"]
	if !ok {
		return false
	}
	if p.hasher == nil {
		return false
	}
	return p.hasher.Check(password, user.Password)
}
