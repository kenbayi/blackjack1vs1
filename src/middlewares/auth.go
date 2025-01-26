package middlewares

import (
	"blackjack/src/db"
	"context"
	"net/http"
)

func ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session token from the cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate session token in Redis
		userID, err := db.RedisClient.Get(r.Context(), cookie.Value).Result()
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set userID in context for future use
		r = r.WithContext(context.WithValue(r.Context(), "user_id", userID))
		next.ServeHTTP(w, r)
	})
}
