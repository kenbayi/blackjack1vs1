package model

import "time"

type UserProfile struct {
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
