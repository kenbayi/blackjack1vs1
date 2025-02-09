package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetRooms retrieves active game rooms from in-memory storage
func (h *Hub) GetRooms(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var rooms []map[string]interface{}

	for roomID, room := range h.Rooms {
		rooms = append(rooms, map[string]interface{}{
			"room_id": roomID,
			"players": len(room.Players),
		})
	}

	// If no rooms are available, return a 404 response
	if len(rooms) == 0 {
		http.Error(w, `{"error": "No active rooms found"}`, http.StatusNotFound)
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
