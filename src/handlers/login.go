package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"sync"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginMutex sync.Mutex

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Lock the critical section
	loginMutex.Lock()
	defer loginMutex.Unlock()

	var req LoginRequest

	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := models.GetUserByUsername(db.PostgresDB, req.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	log.Printf("Logging user: %s", req.Username)

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate session token
	sessionToken := uuid.New().String()

	// Store session in Redis
	err = db.RedisClient.Set(r.Context(), sessionToken, user.ID, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	log.Printf("Logged user: %s", req.Username)
	w.Write([]byte(`{"message": "Login successful"}`))
}
