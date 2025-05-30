package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"user_svc/internal/adapter/postgres/dao"
	"user_svc/internal/model"
	"user_svc/pkg/postgres"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, customer model.User) error {
	query := `
		INSERT INTO users (
			id, username, email, 
			created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.Username,
		customer.Email,
		customer.CreatedAt,
		customer.UpdatedAt,
		customer.IsDeleted,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique_customer_email") {
			return model.ErrEmailAlreadyRegistered
		}
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *UserRepository) PatchByID(ctx context.Context, userUpdated *model.UserUpdateData) error {
	if userUpdated.ID == nil {
		return fmt.Errorf("missing user ID for patch")
	}

	query := `UPDATE users SET`
	args := []interface{}{}
	argID := 1
	setClauses := []string{}

	if userUpdated.Username != nil && *userUpdated.Username != "" {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argID))
		args = append(args, *userUpdated.Username)
		argID++
	}
	if userUpdated.Nickname != nil && *userUpdated.Nickname != "" {
		setClauses = append(setClauses, fmt.Sprintf("nickname = $%d", argID))
		args = append(args, *userUpdated.Nickname)
		argID++
	}
	if userUpdated.Rating != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating = $%d", argID))
		args = append(args, *userUpdated.Rating)
		argID++
	}
	if userUpdated.Bio != nil && *userUpdated.Bio != "" {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argID))
		args = append(args, *userUpdated.Bio)
		argID++
	}
	if userUpdated.Balance != nil {
		setClauses = append(setClauses, fmt.Sprintf("balance = $%d", argID))
		args = append(args, *userUpdated.Balance)
		argID++
	}
	if userUpdated.IsDeleted != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_deleted = $%d", argID))
		args = append(args, *userUpdated.IsDeleted)
		argID++
	}
	if userUpdated.UpdatedAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argID))
		args = append(args, *userUpdated.UpdatedAt)
		argID++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query += " " + strings.Join(setClauses, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", argID)
	args = append(args, *userUpdated.ID)

	if tx, ok := postgres.TxFromCtx(ctx); ok {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("patch failed in transaction: %w", err)
		}
	} else {
		_, err := r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("patch failed: %w", err)
		}
	}

	return nil
}

func (r *UserRepository) GetWithFilter(ctx context.Context, filter model.UserFilter) (model.User, error) {
	whereClause, args := dao.FromUserFilter(filter)
	query := `
		SELECT 
			id, username, nickname, role, bio, email,
			created_at, updated_at, is_deleted, rating, balance
		FROM users
	` + whereClause

	var customerDAO dao.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&customerDAO.ID,
		&customerDAO.Username,
		&customerDAO.Nickname,
		&customerDAO.Role,
		&customerDAO.Bio,
		&customerDAO.Email,
		&customerDAO.CreatedAt,
		&customerDAO.UpdatedAt,
		&customerDAO.IsDeleted,
		&customerDAO.Rating,
		&customerDAO.Balance,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrNotFound
		}
		return model.User{}, fmt.Errorf("failed to get customer: %w", err)
	}

	return dao.ToUser(customerDAO), nil
}

func (r *UserRepository) GetListWithFilter(ctx context.Context, filter model.UserFilter) ([]model.User, error) {
	whereClause, args := dao.FromUserFilter(filter)
	query := `
		SELECT 
			id, username, nickname, role, bio, email,
			created_at, updated_at, is_deleted, rating, balance
		FROM users
	` + whereClause

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []model.User
	for rows.Next() {
		var customerDAO dao.User
		err := rows.Scan(
			&customerDAO.ID,
			&customerDAO.Username,
			&customerDAO.Nickname,
			&customerDAO.Role,
			&customerDAO.Bio,
			&customerDAO.Email,
			&customerDAO.CreatedAt,
			&customerDAO.UpdatedAt,
			&customerDAO.IsDeleted,
			&customerDAO.Rating,
			&customerDAO.Balance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, dao.ToUser(customerDAO))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return customers, nil
}

func (r *UserRepository) GetBalance(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT balance FROM users WHERE id = $1`

	var balance int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID int64, newBalance int64) error {
	query := `UPDATE users SET balance = $1, updated_at = NOW() WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *UserRepository) GetRating(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT rating FROM users WHERE id = $1`

	var rating sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&rating)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}
		return 0, fmt.Errorf("failed to get rating: %w", err)
	}

	if !rating.Valid {
		return 0, nil
	}

	return rating.Int64, nil
}

func (r *UserRepository) UpdateRating(ctx context.Context, userID int64, newRating int64) error {
	query := `UPDATE users SET rating = $1, updated_at = NOW() WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, newRating, userID)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrNotFound
	}

	return nil
}
