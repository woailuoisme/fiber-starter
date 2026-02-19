package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	database "fiber-starter/internal/db"
	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/platform/helpers"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	psqlsm "github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/sqlite"
	sqlitesm "github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/scan"
	"go.uber.org/zap"
)

const (
	dialectPostgres = "psql"
	dialectSQLite   = "sqlite"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUsers(page, limit int) ([]models.User, int64, error)
	UpdateUser(id int64, updates map[string]interface{}) error
	DeleteUser(id int64) error
	UpdateProfile(id int64, profile *models.User) error
	SearchUsers(query string, page, limit int) ([]models.User, int64, error)
}

// userService 用户服务实现
type userService struct {
	db *database.Connection
}

// NewUserService 创建用户服务实例
func NewUserService(db *database.Connection) UserService {
	return &userService{
		db: db,
	}
}

// GetUserByID Get user by ID
func (s *userService) GetUserByID(id int64) (*models.User, error) {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return nil, err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return nil, err
	}

	q, err := userSelectOneByIDQuery(dialect, id)
	if err != nil {
		return nil, err
	}

	u, err := bob.One(ctx, bob.NewDB(db), q, scan.StructMapper[models.User]())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &u, nil
}

// GetUserByEmail Get user by email
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return nil, err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return nil, err
	}

	q, err := userSelectOneByEmailQuery(dialect, email)
	if err != nil {
		return nil, err
	}

	u, err := bob.One(ctx, bob.NewDB(db), q, scan.StructMapper[models.User]())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &u, nil
}

// GetUsers Get user list (paginated)
func (s *userService) GetUsers(page, limit int) ([]models.User, int64, error) {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return nil, 0, err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := int64((page - 1) * limit)

	countQ, listQ, err := userListQueries(dialect, int64(limit), offset, "")
	if err != nil {
		return nil, 0, err
	}

	total, err := bob.One(ctx, bob.NewDB(db), countQ, scan.SingleColumnMapper[int64])
	if err != nil {
		helpers.LogError("Failed to get user count", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	users, err := bob.All(ctx, bob.NewDB(db), listQ, scan.StructMapper[models.User]())
	if err != nil {
		helpers.LogError("Failed to get user list", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get user list: %w", err)
	}

	return users, total, nil
}

// UpdateUser Update user information
func (s *userService) UpdateUser(id int64, updates map[string]interface{}) error {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates)+1)
	args := make([]any, 0, len(updates)+2)

	for k, v := range updates {
		switch strings.ToLower(strings.TrimSpace(k)) {
		case "name", "email", "avatar", "phone", "status", "email_verified_at":
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", k))
			args = append(args, v)
		default:
		}
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC())
	args = append(args, id)

	q, err := userUpdateRawQuery(dialect, strings.Join(setClauses, ", "), args...)
	if err != nil {
		return err
	}

	if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
		helpers.LogError("Failed to update user", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser Delete user (soft delete)
func (s *userService) DeleteUser(id int64) error {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	q, err := userSoftDeleteQuery(dialect, id, now)
	if err != nil {
		return err
	}

	if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
		helpers.LogError("Failed to delete user", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateProfile Update user profile
func (s *userService) UpdateProfile(id int64, profile *models.User) error {
	if profile == nil {
		return nil
	}

	updates := make(map[string]interface{})
	if profile.Name != "" {
		updates["name"] = profile.Name
	}
	if profile.Avatar != nil {
		updates["avatar"] = profile.Avatar
	}
	if profile.Phone != nil {
		updates["phone"] = profile.Phone
	}

	return s.UpdateUser(id, updates)
}

// SearchUsers Search users
func (s *userService) SearchUsers(query string, page, limit int) ([]models.User, int64, error) {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return nil, 0, err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := int64((page - 1) * limit)
	search := strings.TrimSpace(query)

	countQ, listQ, err := userListQueries(dialect, int64(limit), offset, search)
	if err != nil {
		return nil, 0, err
	}

	total, err := bob.One(ctx, bob.NewDB(db), countQ, scan.SingleColumnMapper[int64])
	if err != nil {
		helpers.LogError("Failed to get search result count", zap.Error(err), zap.String("query", query))
		return nil, 0, fmt.Errorf("failed to get search result count: %w", err)
	}

	users, err := bob.All(ctx, bob.NewDB(db), listQ, scan.StructMapper[models.User]())
	if err != nil {
		helpers.LogError("Failed to search users", zap.Error(err), zap.String("query", query))
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

func userColumns() []any {
	return []any{
		"id", "name", "email", "password", "avatar", "phone", "status",
		"email_verified_at", "created_at", "updated_at", "deleted_at",
	}
}

func userSelectOneByIDQuery(dialect string, id int64) (bob.Query, error) {
	switch dialect {
	case dialectPostgres:
		return psql.Select(
			psqlsm.Columns(userColumns()...),
			psqlsm.From("users"),
			psqlsm.Where(psql.Raw("id = ?", id)),
			psqlsm.Limit(1),
		), nil
	case dialectSQLite:
		return sqlite.Select(
			sqlitesm.Columns(userColumns()...),
			sqlitesm.From("users"),
			sqlitesm.Where(sqlite.Raw("id = ?", id)),
			sqlitesm.Limit(1),
		), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func userSelectOneByEmailQuery(dialect string, email string) (bob.Query, error) {
	switch dialect {
	case dialectPostgres:
		return psql.Select(
			psqlsm.Columns(userColumns()...),
			psqlsm.From("users"),
			psqlsm.Where(psql.Raw("email = ? AND deleted_at IS NULL", email)),
			psqlsm.Limit(1),
		), nil
	case dialectSQLite:
		return sqlite.Select(
			sqlitesm.Columns(userColumns()...),
			sqlitesm.From("users"),
			sqlitesm.Where(sqlite.Raw("email = ? AND deleted_at IS NULL", email)),
			sqlitesm.Limit(1),
		), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func userListQueries(dialect string, limit int64, offset int64, search string) (bob.Query, bob.Query, error) {
	baseWhere := "deleted_at IS NULL"
	args := []any{}

	if search != "" {
		baseWhere = "deleted_at IS NULL AND (name LIKE ? OR email LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	switch dialect {
	case dialectPostgres:
		countQ := psql.Select(
			psqlsm.Columns("COUNT(*)"),
			psqlsm.From("users"),
			psqlsm.Where(psql.Raw(baseWhere, args...)),
		)
		listQ := psql.Select(
			psqlsm.Columns(userColumns()...),
			psqlsm.From("users"),
			psqlsm.Where(psql.Raw(baseWhere, args...)),
			psqlsm.OrderBy(psql.Raw("created_at DESC")),
			psqlsm.Limit(limit),
			psqlsm.Offset(offset),
		)
		return countQ, listQ, nil
	case dialectSQLite:
		countQ := sqlite.Select(
			sqlitesm.Columns("COUNT(*)"),
			sqlitesm.From("users"),
			sqlitesm.Where(sqlite.Raw(baseWhere, args...)),
		)
		listQ := sqlite.Select(
			sqlitesm.Columns(userColumns()...),
			sqlitesm.From("users"),
			sqlitesm.Where(sqlite.Raw(baseWhere, args...)),
			sqlitesm.OrderBy(sqlite.Raw("created_at DESC")),
			sqlitesm.Limit(limit),
			sqlitesm.Offset(offset),
		)
		return countQ, listQ, nil
	default:
		return nil, nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func userUpdateRawQuery(dialect string, setClause string, args ...any) (bob.Query, error) {
	switch dialect {
	case dialectPostgres:
		return psql.RawQuery("UPDATE users SET "+setClause+" WHERE id = ?", args...), nil
	case dialectSQLite:
		return sqlite.RawQuery("UPDATE users SET "+setClause+" WHERE id = ?", args...), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func userSoftDeleteQuery(dialect string, id int64, deletedAt time.Time) (bob.Query, error) {
	switch dialect {
	case dialectPostgres:
		return psql.RawQuery("UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", deletedAt, deletedAt, id), nil
	case dialectSQLite:
		return sqlite.RawQuery("UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", deletedAt, deletedAt, id), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}
