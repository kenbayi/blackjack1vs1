package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// Global Hub instance
var HubInstance = NewHub()

func init() {
	go HubInstance.Run() // Start the hub
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	playerID, ok := r.Context().Value("user_id").(string)
	if !ok {
		log.Println("PlayerID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		conn.Close()
		return
	}

	// Register the connection with the hub
	HubInstance.Register <- conn

	defer func() {
		HubInstance.Unregister <- conn
		conn.Close()
	}()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}

		if content, ok := msg["content"].(map[string]interface{}); ok {
			content["playerID"] = playerID
		} else {
			msg["content"] = map[string]interface{}{
				"playerID": playerID,
			}
		}

		HubInstance.Messages <- Message{
			Type:    msg["type"].(string),
			Content: msg["content"],
			Conn:    conn,
		}
	}
}
