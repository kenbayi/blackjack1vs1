package model

import "time"

const (
	CustomerRole = "customer"
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
type UserUpdateData struct {
	ID           *int64
	Username     *string
	Email        *string
	PasswordHash *string
	UpdatedAt    *time.Time

	IsDeleted *bool
}
type Token struct {
	AccessToken  string
	RefreshToken string
}
type RequestToChange struct {
	Token       string
	NewPassword string
}
