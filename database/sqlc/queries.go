package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	models "fiber-starter/app/Models"
)

// DBTX matches the database/sql interfaces used by sqlc generated code.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Queries provides a small sqlc-style query layer for user data.
type Queries struct {
	db      DBTX
	dialect string
}

// New creates a query helper bound to a database connection.
func New(db DBTX, dialect string) *Queries {
	return &Queries{db: db, dialect: normalizeDialect(dialect)}
}

// WithTx binds the query helper to a transaction.
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx, dialect: q.dialect}
}

type CreateUserParams struct {
	Name            string
	Email           string
	Password        string
	Avatar          *string
	Phone           *string
	Status          models.UserStatus
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpdateUserParams struct {
	ID              int64
	Name            string
	Email           string
	Password        string
	Avatar          *string
	Phone           *string
	Status          models.UserStatus
	EmailVerifiedAt *time.Time
	UpdatedAt       time.Time
}

type UpdatePasswordParams struct {
	ID        int64
	Password  string
	UpdatedAt time.Time
}

type ResetPasswordByEmailParams struct {
	Email     string
	Password  string
	UpdatedAt time.Time
}

type SoftDeleteUserParams struct {
	ID        int64
	DeletedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	return q.countUsers(ctx, "")
}

func (q *Queries) CountUsersBySearch(ctx context.Context, search string) (int64, error) {
	return q.countUsers(ctx, search)
}

func (q *Queries) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM users WHERE email = "+q.placeholder(1)+" AND deleted_at IS NULL", email)

	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = ` + q.placeholder(1) + ` AND deleted_at IS NULL LIMIT 1`
	return q.scanUser(q.db.QueryRowContext(ctx, query, email))
}

func (q *Queries) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = ` + q.placeholder(1) + ` AND deleted_at IS NULL LIMIT 1`
	return q.scanUser(q.db.QueryRowContext(ctx, query, id))
}

func (q *Queries) ListUsers(ctx context.Context, search string, limit, offset int) ([]models.User, error) {
	args := make([]any, 0, 3)
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(userColumns)
	builder.WriteString(" FROM users WHERE deleted_at IS NULL")

	if normalized := strings.TrimSpace(search); normalized != "" {
		builder.WriteString(" AND (LOWER(name) LIKE LOWER(")
		builder.WriteString(q.placeholder(len(args) + 1))
		builder.WriteString(") OR LOWER(email) LIKE LOWER(")
		builder.WriteString(q.placeholder(len(args) + 2))
		builder.WriteString("))")
		pattern := "%" + normalized + "%"
		args = append(args, pattern, pattern)
	}

	builder.WriteString(" ORDER BY created_at DESC LIMIT ")
	builder.WriteString(q.placeholder(len(args) + 1))
	builder.WriteString(" OFFSET ")
	builder.WriteString(q.placeholder(len(args) + 2))
	args = append(args, limit, offset)

	rows, err := q.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	users := make([]models.User, 0)
	for rows.Next() {
		user, err := q.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (models.User, error) {
	query := `INSERT INTO users (
		name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at
	) VALUES (
		` + q.placeholder(1) + `,
		` + q.placeholder(2) + `,
		` + q.placeholder(3) + `,
		` + q.placeholder(4) + `,
		` + q.placeholder(5) + `,
		` + q.placeholder(6) + `,
		` + q.placeholder(7) + `,
		` + q.placeholder(8) + `,
		` + q.placeholder(9) + `
	) RETURNING ` + userColumns

	return q.scanUser(q.db.QueryRowContext(ctx, query,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.Avatar,
		arg.Phone,
		string(arg.Status),
		arg.EmailVerifiedAt,
		arg.CreatedAt,
		arg.UpdatedAt,
	))
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	query := `UPDATE users SET
		name = ` + q.placeholder(1) + `,
		email = ` + q.placeholder(2) + `,
		password = ` + q.placeholder(3) + `,
		avatar = ` + q.placeholder(4) + `,
		phone = ` + q.placeholder(5) + `,
		status = ` + q.placeholder(6) + `,
		email_verified_at = ` + q.placeholder(7) + `,
		updated_at = ` + q.placeholder(8) + `
	WHERE id = ` + q.placeholder(9) + ` AND deleted_at IS NULL
	RETURNING id`

	var id int64
	return q.db.QueryRowContext(ctx, query,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.Avatar,
		arg.Phone,
		string(arg.Status),
		arg.EmailVerifiedAt,
		arg.UpdatedAt,
		arg.ID,
	).Scan(&id)
}

func (q *Queries) UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error {
	query := `UPDATE users SET password = ` + q.placeholder(1) + `, updated_at = ` + q.placeholder(2) + ` WHERE id = ` + q.placeholder(3) + ` AND deleted_at IS NULL RETURNING id`
	var id int64
	return q.db.QueryRowContext(ctx, query, arg.Password, arg.UpdatedAt, arg.ID).Scan(&id)
}

func (q *Queries) ResetPasswordByEmail(ctx context.Context, arg ResetPasswordByEmailParams) error {
	query := `UPDATE users SET password = ` + q.placeholder(1) + `, updated_at = ` + q.placeholder(2) + ` WHERE email = ` + q.placeholder(3) + ` AND deleted_at IS NULL RETURNING id`
	var id int64
	return q.db.QueryRowContext(ctx, query, arg.Password, arg.UpdatedAt, arg.Email).Scan(&id)
}

func (q *Queries) SoftDeleteUser(ctx context.Context, arg SoftDeleteUserParams) error {
	query := `UPDATE users SET deleted_at = ` + q.placeholder(1) + `, updated_at = ` + q.placeholder(2) + ` WHERE id = ` + q.placeholder(3) + ` AND deleted_at IS NULL RETURNING id`
	var id int64
	return q.db.QueryRowContext(ctx, query, arg.DeletedAt, arg.UpdatedAt, arg.ID).Scan(&id)
}

func (q *Queries) DeleteAllUsers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM users")
	return err
}

func (q *Queries) countUsers(ctx context.Context, search string) (int64, error) {
	query := "SELECT COUNT(1) FROM users WHERE deleted_at IS NULL"
	args := make([]any, 0, 2)
	if normalized := strings.TrimSpace(search); normalized != "" {
		query += " AND (LOWER(name) LIKE LOWER(" + q.placeholder(1) + ") OR LOWER(email) LIKE LOWER(" + q.placeholder(2) + "))"
		pattern := "%" + normalized + "%"
		args = append(args, pattern, pattern)
	}

	row := q.db.QueryRowContext(ctx, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (q *Queries) scanUser(scanner interface{ Scan(...any) error }) (models.User, error) {
	var user models.User
	var avatar sql.NullString
	var phone sql.NullString
	var status string
	var verifiedAt sql.NullTime
	var deletedAt sql.NullTime

	if err := scanner.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&avatar,
		&phone,
		&status,
		&verifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	); err != nil {
		return models.User{}, err
	}

	if avatar.Valid {
		user.Avatar = &avatar.String
	}
	if phone.Valid {
		user.Phone = &phone.String
	}
	if verifiedAt.Valid {
		value := verifiedAt.Time
		user.EmailVerifiedAt = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time
		user.DeletedAt = &value
	}
	user.Status = models.UserStatus(status)
	return user, nil
}

func (q *Queries) placeholder(index int) string {
	if q.dialect == "sqlite" {
		return "?"
	}
	if index < 1 {
		index = 1
	}
	return fmt.Sprintf("$%d", index)
}

func normalizeDialect(dialect string) string {
	switch strings.ToLower(strings.TrimSpace(dialect)) {
	case "sqlite", "sqlite3":
		return "sqlite"
	case "psql", "postgres", "postgresql":
		return "postgres"
	default:
		return "postgres"
	}
}

const userColumns = "id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at"
