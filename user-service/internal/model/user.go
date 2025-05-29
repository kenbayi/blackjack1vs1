package model

import "time"

const (
	UserRole  = "user"
	AdminRole = "admin"
)

type User struct {
	ID        int64
	Username  string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
	Nickname  *string
	Bio       *string
	Balance   *int64
	Rating    *int64
}

type UserFilter struct {
	ID        *int64
	Username  *string
	Nickname  *string
	Balance   *int64
	Rating    *int64
	Email     *string
	IsDeleted *bool
}

type UserUpdateData struct {
	ID        *int64
	Username  *string
	Email     *string
	UpdatedAt *time.Time
	Bio       *string
	Nickname  *string
	Balance   *int64
	Rating    *int64
	IsDeleted *bool
}
