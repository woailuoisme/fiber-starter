package auth

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type AuthRepository struct {
	db bun.IDB
}

func NewAuthRepository(db bun.IDB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateOTP(ctx context.Context, otp *AuthOTP) error {
	_, err := r.db.NewInsert().Model(otp).Returning("*").Exec(ctx)
	return err
}

func (r *AuthRepository) GetLatestOTPByEmailPurpose(ctx context.Context, email string, purpose AuthOTPPurpose) (*AuthOTP, error) {
	var otp AuthOTP
	err := r.db.NewSelect().
		Model(&otp).
		Where("email = ?", email).
		Where("purpose = ?", purpose).
		Where("consumed_at IS NULL").
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *AuthRepository) IncrementOTPAttempts(ctx context.Context, id int64, updatedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*AuthOTP)(nil)).
		Set("attempts = attempts + 1").
		Set("updated_at = ?", updatedAt).
		Where("id = ? AND consumed_at IS NULL", id).
		Exec(ctx)
	return err
}

func (r *AuthRepository) ConsumeOTP(ctx context.Context, id int64, consumedAt, updatedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*AuthOTP)(nil)).
		Set("consumed_at = ?", consumedAt).
		Set("updated_at = ?", updatedAt).
		Where("id = ? AND consumed_at IS NULL", id).
		Exec(ctx)
	return err
}

func (r *AuthRepository) DeleteOTPByID(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*AuthOTP)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
