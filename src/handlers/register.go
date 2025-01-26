package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var registerMutex sync.Mutex

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		http.Error(w, "Username must be at least 3 characters and password at least 6 characters", http.StatusBadRequest)
		return
	}

	log.Println("Registering user: ", req.Username)
	// Lock the critical section
	registerMutex.Lock()
	defer registerMutex.Unlock()

	exists, err := models.CheckUsernameExists(db.PostgresDB, req.Username)
	if err != nil {
		http.Error(w, "Failed to check username", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	userID, err := models.InsertUser(db.PostgresDB, req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("Registered user: ", userID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user_id": userID,
	})
}