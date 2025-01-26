package handlers

import (
	"blackjack/src/db"
	"log"
	"net/http"
	"sync"
	"time"
)

var logoutMutex sync.Mutex

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Lock the critical section
	logoutMutex.Lock()
	defer logoutMutex.Unlock()

	// Get the session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session token found", http.StatusUnauthorized)
		return
	}

	// Delete the session token from Redis
	userID, err := db.RedisClient.Get(r.Context(), cookie.Value).Result()
	err = db.RedisClient.Del(r.Context(), cookie.Value).Err()
	if err != nil {
		http.Error(w, "Failed to log out", http.StatusInternalServerError)
		return
	}
	log.Printf("logged out user with ID: %s", userID)

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logout successful"}`))
}
