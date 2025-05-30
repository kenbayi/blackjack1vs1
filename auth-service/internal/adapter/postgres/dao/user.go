package dao

import (
	"auth_svc/internal/model"
	"strconv"
	"strings"
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	IsDeleted    bool      `db:"is_deleted"`
}

func ToUser(customer User) model.User {
	return model.User{
		ID:           customer.ID,
		Username:     customer.Username,
		Email:        customer.Email,
		PasswordHash: customer.PasswordHash,
		CreatedAt:    customer.CreatedAt,
		UpdatedAt:    customer.UpdatedAt,
		IsDeleted:    customer.IsDeleted,
	}
}

func FromUserFilter(filter model.UserFilter) (string, []interface{}) {
	var query string
	var args []interface{}
	var conditions []string
	argCounter := 1

	if filter.ID != nil {
		conditions = append(conditions, "id = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.ID)
		argCounter++
	}

	if filter.Username != nil {
		conditions = append(conditions, "name = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Username)
		argCounter++
	}

	if filter.Email != nil {
		conditions = append(conditions, "email = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Email)
		argCounter++
	}

	if filter.IsDeleted != nil {
		conditions = append(conditions, "is_deleted = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.IsDeleted)
		argCounter++
	}

	if len(conditions) > 0 {
		query = " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}
