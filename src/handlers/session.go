package handlers

import (
	"encoding/json"
	"net/http"
)

// SessionHandler validates the session and returns user details.
func SessionHandler(w http.ResponseWriter, r *http.Request) {
	// The ValidateSession middleware sets "user_id" in context.
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Assuming user_id is stored as a string in context.
	userIDStr, ok := userIDVal.(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Simply return the user ID.
	response := map[string]interface{}{
		"user_id": userIDStr,
		"message": "Session is valid",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
