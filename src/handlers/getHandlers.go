package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

// GetRooms retrieves active game rooms from in-memory storage
func (h *Hub) GetRooms(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var rooms []map[string]interface{}

	// Fetch all room keys from Redis
	roomKeys, err := db.RedisClient.Keys(db.Ctx, "room:*").Result()
	if err != nil {
		log.Println("Error fetching room keys from Redis:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, roomKey := range roomKeys {
		// Get all room details
		roomData, err := db.RedisClient.HGetAll(db.Ctx, roomKey).Result()
		if err != nil {
			log.Println("Error fetching room data from Redis for", roomKey, ":", err)
			continue
		}

		// Only add rooms that are "waiting"
		if roomData["status"] == "waiting" && !strings.Contains(roomData["players"], ",") {
			rooms = append(rooms, map[string]interface{}{
				"roomID":  roomData["roomID"],
				"players": []string{roomData["players"]}, // Ensure it's always an array
				"bet":     roomData["bet"],
			})
		}
	}

	// If no rooms are available, return an empty list
	if len(rooms) == 0 {
		log.Println("⚠️ No active rooms, returning empty list.")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// GET /history/{id} → Returns game history
func GetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"] //extract 'id' from the URL

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	historyChannel := make(chan *models.GameRoom)
	errChannel := make(chan error)

	go func() {
		historyData, err := models.GetHistory(db.PostgresDB, userID)
		if err != nil {
			errChannel <- err
			return
		}
		if historyData == nil {
			errChannel <- fmt.Errorf("History not found")
			return
		}
		historyChannel <- historyData
	}()

	select {
	case history := <-historyChannel:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	case err := <-errChannel:
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

// GET /user/{username} → Returns user details by username
func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"] // Extracts 'username' from the URL

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	userChannel := make(chan *models.User)
	errChannel := make(chan error)

	go func() {
		user, err := models.GetUserByUsername(db.PostgresDB, username)
		if err != nil {
			errChannel <- err
			return
		}
		if user == nil {
			errChannel <- fmt.Errorf("User not found")
			return
		}
		userChannel <- user
	}()

	select {
	case user := <-userChannel:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	case err := <-errChannel:
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
