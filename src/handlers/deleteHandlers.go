package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Call DeleteUser function
	err = models.DeleteUser(db.PostgresDB, userID)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}
