package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`         // Unique identifier
	Username  string    `json:"username"`   // Username
	Password  string    `json:"-"`          // Hashed password (not exposed in JSON)
	CreatedAt time.Time `json:"created_at"` // Account creation timestamp
}
