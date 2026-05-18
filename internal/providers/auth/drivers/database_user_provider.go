package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	exceptions "fiber-starter/internal/common/exceptions"
	database "fiber-starter/internal/providers/database/contracts"
	hashContracts "fiber-starter/internal/providers/hash/contracts"

	"github.com/uptrace/bun"
)

// AuthUser represents the database schema mapped for user authentication
type AuthUser struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            int64  `bun:"id,pk,autoincrement"`
	Name          string `bun:"name"`
	Email         string `bun:"email,unique"`
	Password      string `bun:"password"`
}

// DatabaseUserProvider implements the UserProvider interface using the application database
type DatabaseUserProvider struct {
	db           database.Connection
	table        string
	hasher       hashContracts.Hasher
	modelCreator func() any
}

// NewDatabaseUserProvider creates a new database user provider instance
func NewDatabaseUserProvider(db database.Connection, table string, hasher hashContracts.Hasher) *DatabaseUserProvider {
	return &DatabaseUserProvider{db: db, table: table, hasher: hasher}
}

// SetModelCreator sets the model creator function
func (p *DatabaseUserProvider) SetModelCreator(creator func() any) {
	if p != nil {
		p.modelCreator = creator
	}
}

// RetrieveById retrieves a user by their unique identifier
func (p *DatabaseUserProvider) RetrieveById(id int64) (any, error) {
	if p == nil || p.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", fmt.Errorf("database connection not initialized"))
	}

	if p.table == "users" {
		bunDB, err := p.db.BunDB()
		if err != nil {
			return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
		}
		if p.modelCreator != nil {
			usr := p.modelCreator()
			err = bunDB.NewSelect().Model(usr).Where("id = ?", id).Where("deleted_at IS NULL").Scan(context.Background())
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, nil
				}
				return nil, err
			}
			return usr, nil
		}
		var usr AuthUser
		err = bunDB.NewSelect().Model(&usr).Where("id = ?", id).Where("deleted_at IS NULL").Scan(context.Background())
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &usr, nil
	}

	// Generic raw SQL for other tables
	db, err := p.db.GetDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}

	// nolint:gosec // table name is from trusted config
	query := fmt.Sprintf("SELECT id, name, email, password FROM %s WHERE id = $1 AND deleted_at IS NULL LIMIT 1", p.table)
	row := db.QueryRowContext(context.Background(), query, id)

	var usr AuthUser
	err = row.Scan(&usr.ID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}

// RetrieveByCredentials retrieves a user by the given credentials (e.g. email)
func (p *DatabaseUserProvider) RetrieveByCredentials(credentials map[string]string) (any, error) {
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
		if p.modelCreator != nil {
			usr := p.modelCreator()
			err = bunDB.NewSelect().Model(usr).Where("email = ?", email).Where("deleted_at IS NULL").Scan(context.Background())
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, nil
				}
				return nil, err
			}
			return usr, nil
		}
		var usr AuthUser
		err = bunDB.NewSelect().Model(&usr).Where("email = ?", email).Where("deleted_at IS NULL").Scan(context.Background())
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &usr, nil
	}

	// Generic raw SQL for other tables
	db, err := p.db.GetDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}

	// nolint:gosec // table name is from trusted config
	query := fmt.Sprintf("SELECT id, name, email, password FROM %s WHERE email = $1 AND deleted_at IS NULL LIMIT 1", p.table)
	row := db.QueryRowContext(context.Background(), query, email)

	var usr AuthUser
	err = row.Scan(&usr.ID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}

// ValidateCredentials validates a user against the given credentials
func (p *DatabaseUserProvider) ValidateCredentials(user any, credentials map[string]string) bool {
	password, ok := credentials["password"]
	if !ok {
		return false
	}
	if p.hasher == nil {
		return false
	}

	var userPassword string
	if u, ok := user.(*AuthUser); ok {
		userPassword = u.Password
	} else {
		userPassword = getLocalStringField(user, "Password")
	}

	return p.hasher.Check(password, userPassword)
}

func getLocalStringField(obj any, name string) string {
	if obj == nil {
		return ""
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}
	f := val.FieldByName(name)
	if !f.IsValid() {
		return ""
	}
	if f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}
