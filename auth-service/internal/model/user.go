package model

import "time"

const (
	UserRole  = "user"
	AdminRole = "admin"
)

type User struct {
	ID              int64
	Username        string
	Email           string
	Role            string
	CurrentPassword string
	NewPassword     string
	PasswordHash    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	IsDeleted       bool
}

type UserFilter struct {
	ID       *int64
	Username *string
	Email    *string

	IsDeleted *bool
}

type UserUpdateData struct {
	ID           *int64
	Username     *string
	Email        *string
	PasswordHash *string
	UpdatedAt    *time.Time

	IsDeleted *bool
}
