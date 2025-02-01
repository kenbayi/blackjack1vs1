package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Hub instance
var hub = NewHub()

func init() {
	// Start the hub in a goroutine
	go hub.Run()
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
	hub.Register <- conn

	defer func() {
		hub.Unregister <- conn
		conn.Close()
	}()

	// Listen for messages
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
			// If content does not exist, create it and add playerID
			msg["content"] = map[string]interface{}{
				"playerID": playerID,
			}
		}

		// Send the message to the hub
		hub.Messages <- Message{
			Type:    msg["type"].(string),
			Content: msg["content"],
			Conn:    conn,
		}
	}
}
