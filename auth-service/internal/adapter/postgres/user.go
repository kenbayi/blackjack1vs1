package postgres

import (
	"auth_svc/internal/adapter/postgres/dao"
	"auth_svc/internal/model"
	"auth_svc/pkg/postgres"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
			username, email, password_hash, 
			created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		customer.Username,
		customer.Email,
		customer.PasswordHash,
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

	if userUpdated.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argID))
		args = append(args, *userUpdated.Username)
		argID++
	}
	if userUpdated.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argID))
		args = append(args, *userUpdated.Email)
		argID++
	}
	if userUpdated.PasswordHash != nil {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argID))
		args = append(args, *userUpdated.PasswordHash)
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
			id, username, email, password_hash,
			created_at, updated_at, is_deleted
		FROM users
	` + whereClause

	var customerDAO dao.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&customerDAO.ID,
		&customerDAO.Username,
		&customerDAO.Email,
		&customerDAO.PasswordHash,
		&customerDAO.CreatedAt,
		&customerDAO.UpdatedAt,
		&customerDAO.IsDeleted,
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
			id, username, email, password_hash,
			created_at, updated_at, is_deleted
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
			&customerDAO.Email,
			&customerDAO.PasswordHash,
			&customerDAO.CreatedAt,
			&customerDAO.UpdatedAt,
			&customerDAO.IsDeleted,
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
