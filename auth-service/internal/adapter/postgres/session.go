package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"auth_svc/internal/adapter/postgres/dao"
	"auth_svc/internal/model"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, session model.Session) error {
	query := `
		INSERT INTO refresh_tokens (
			user_id, refresh_token, 
			created_at, expires_at
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.UserID,
		session.RefreshToken,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (model.Session, error) {
	query := `
		SELECT 
			user_id, refresh_token,
			created_at, expires_at
		FROM refresh_tokens
		WHERE refresh_token = $1
	`

	var sessionDAO dao.Session
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&sessionDAO.UserID,
		&sessionDAO.RefreshToken,
		&sessionDAO.CreatedAt,
		&sessionDAO.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Session{}, model.ErrNotFound
		}
		return model.Session{}, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return dao.ToSession(sessionDAO), nil
}

func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE refresh_token = $1
	`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return model.ErrNotFound
	}

	return nil
}
