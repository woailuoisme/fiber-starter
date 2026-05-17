package user

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type UserRepository struct {
	db bun.IDB
}

func NewUserRepository(db bun.IDB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.selectUsers().Count(ctx)
	return int64(count), err
}

func (r *UserRepository) CountBySearch(ctx context.Context, search string) (int64, error) {
	sel := r.selectUsers()
	if pattern := searchPattern(search); pattern != "" {
		sel = sel.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereOr("LOWER(name) LIKE LOWER(?)", pattern).
				WhereOr("LOWER(email) LIKE LOWER(?)", pattern)
		})
	}

	count, err := sel.Count(ctx)
	return int64(count), err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.selectUsers().Where("email = ?", email).Exists(ctx)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.selectUsers().
		Where("email = ?", email).
		Limit(1).
		Scan(ctx, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := r.selectUsers().
		Where("id = ?", id).
		Limit(1).
		Scan(ctx, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, search string, limit, offset int) ([]User, error) {
	if limit < 1 {
		limit = 15
	}
	if offset < 0 {
		offset = 0
	}

	sel := r.selectUsers()
	if pattern := searchPattern(search); pattern != "" {
		sel = sel.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereOr("LOWER(name) LIKE LOWER(?)", pattern).
				WhereOr("LOWER(email) LIKE LOWER(?)", pattern)
		})
	}

	var users []User
	if err := sel.OrderExpr("created_at DESC").Limit(limit).Offset(offset).Scan(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	_, err := r.db.NewInsert().Model(user).Returning("*").Exec(ctx)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		Column("name", "email", "password", "avatar", "phone", "status", "email_verified_at", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, password string, updatedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*User)(nil)).
		Set("password = ?", password).
		Set("updated_at = ?", updatedAt).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	return err
}

func (r *UserRepository) ResetPasswordByEmail(ctx context.Context, email, password string, updatedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*User)(nil)).
		Set("password = ?", password).
		Set("updated_at = ?", updatedAt).
		Where("email = ? AND deleted_at IS NULL", email).
		Exec(ctx)
	return err
}

func (r *UserRepository) SoftDelete(ctx context.Context, id int64, deletedAt, updatedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*User)(nil)).
		Set("deleted_at = ?", deletedAt).
		Set("updated_at = ?", updatedAt).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	return err
}

func (r *UserRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.NewDelete().Model((*User)(nil)).Exec(ctx)
	return err
}

func (r *UserRepository) selectUsers() *bun.SelectQuery {
	return r.db.NewSelect().
		Model((*User)(nil)).
		Where("deleted_at IS NULL")
}

func searchPattern(search string) string {
	search = strings.TrimSpace(search)
	if search == "" {
		return ""
	}
	if strings.Contains(search, "%") || strings.Contains(search, "_") {
		return search
	}
	return "%" + search + "%"
}
