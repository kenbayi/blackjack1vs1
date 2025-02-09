package handlers

import (
	"encoding/json"
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
