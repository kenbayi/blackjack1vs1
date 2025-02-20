package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"encoding/json"
	"log"
	"net/http"
)

// UpdateUserProfile updates the user's username
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	exists, err := models.CheckUsernameExists(db.PostgresDB, req.Username)
	if err != nil {
		log.Println("Failed to check username:", err)
		http.Error(w, "Failed to check username", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Process update synchronously
	err = models.UpdateUsername(db.PostgresDB, req.UserID, req.Username)
	if err != nil {
		log.Println("Failed to update username:", err)
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	log.Printf("User %d updated username to %s", req.UserID, req.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

// UpdateUserBalance updates the user's balance (add money)
func UpdateUserBalance(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process update synchronously to ensure response is correct
	err := models.UpdateUserBalance(db.PostgresDB, req.UserID, req.Amount)
	if err != nil {
		log.Println("Failed to update balance:", err)
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		return
	}

	log.Printf("User %d balance updated by %d", req.UserID, req.Amount)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Balance updated successfully"})
}