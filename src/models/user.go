package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

type User struct {
	ID        int       `json:"id"`         // Unique identifier
	Username  string    `json:"username"`   // Username
	Password  string    `json:"-"`          // Hashed password (not exposed in JSON)
	CreatedAt time.Time `json:"created_at"` // Account creation timestamp
	Balance   int       `json:"balance"`    // User Balance
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	query := `SELECT id, username, password, created_at, balance FROM users WHERE username = $1`
	var user User
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.Balance)
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

func UpdatePlayerBalances(db *sql.DB, ctx context.Context, bet string, winnerID, loserID string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance - $1 WHERE id = $2`, bet, loserID)
	if err != nil {
		return fmt.Errorf("error updating loser balance: %v", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id = $2`, bet, winnerID)
	if err != nil {
		return fmt.Errorf("error updating winner balance: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// UpdateUserBalance updates the user's balance in a transaction-safe manner
func UpdateUserBalance(db *sql.DB, userID int, amount int) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	// Update balance
	_, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id = $2`, amount, userID)
	if err != nil {
		log.Println("Error updating user balance:", err)
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

// UpdateUsername updates a user's username in a transaction-safe manner
func UpdateUsername(db *sql.DB, userID int, newUsername string) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	// Update username
	_, err = tx.ExecContext(ctx, `UPDATE users SET username = $1 WHERE id = $2`, newUsername, userID)
	if err != nil {
		log.Println("Error updating user profile:", err)
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

func DeleteUser(db *sql.DB, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// First, delete any related game sessions or references (if needed)
	_, err = tx.ExecContext(ctx, `DELETE FROM game_rooms WHERE player1_id = $1 OR player2_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("error deleting user game sessions: %v", err)
	}

	// Then, delete the user
	_, err = tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("error deleting user account: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
