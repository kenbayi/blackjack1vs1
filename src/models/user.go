package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type User struct {
	ID        int       `json:"id"`         // Unique identifier
	Username  string    `json:"username"`   // Username
	Password  string    `json:"-"`          // Hashed password (not exposed in JSON)
	CreatedAt time.Time `json:"created_at"` // Account creation timestamp
	Bet       int       `json:"bet"`
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	query := `SELECT id, username, password, created_at FROM users WHERE username = $1`
	var user User
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func CheckUsernameExists(db *sql.DB, username string) (bool, error) {
	query := `SELECT COUNT(1) FROM users WHERE username = $1`
	var count int
	err := db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func InsertUser(db *sql.DB, username string, hashedPassword string) (int, error) {
	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	var userID int
	err := db.QueryRow(query, username, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func GetPlayerBalance(db *sql.DB, playerID string) (int, error) {
	// Convert playerID (string) to an int
	id, err := strconv.Atoi(playerID)
	if err != nil {
		return 0, fmt.Errorf("invalid playerID: %v", err)
	}
	query := `SELECT balance FROM users WHERE id = $1`
	var balance int
	err = db.QueryRow(query, id).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
