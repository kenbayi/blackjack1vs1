package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
	Conn    *websocket.Conn
}

type Room struct {
	ID      string
	Players map[*websocket.Conn]string // Connections mapped to player IDs
}

type Hub struct {
	Clients    map[*websocket.Conn]bool // All connected clients
	Rooms      map[string]*Room         // Active rooms
	Register   chan *websocket.Conn     // Channel for new connections
	Unregister chan *websocket.Conn     // Channel for disconnected clients
	Messages   chan Message             // Incoming messages
	mu         sync.Mutex               // Mutex for safe room updates
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		Rooms:      make(map[string]*Room),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Messages:   make(chan Message),
	}
}

// Run starts the Hub to process connections and messages
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.Clients[conn] = true
			log.Println("Client connected", conn.RemoteAddr())

		case conn := <-h.Unregister:
			if _, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				log.Println("Client disconnected")
			}

		case msg := <-h.Messages:
			h.handleMessage(msg)
		}
	}
}

func (h *Hub) handleMessage(msg Message) {
	switch msg.Type {
	case "create_room":
		h.createRoom(msg)
	case "join_room":
		h.joinRoom(msg)
	case "ready":
		h.handleReady(msg)
	case "leave_room":
		h.leaveRoom(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *Hub) broadcastAll(message map[string]interface{}) {
	// Broadcast to all connected clients globally
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.Clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Error broadcasting message to all: %v", err)
			client.Close()
			delete(h.Clients, client)
		}
	}
}

func (h *Hub) broadcastRoom(roomID string, message map[string]interface{}) {
	// Broadcast only to the clients in a specific room
	h.mu.Lock()
	defer h.mu.Unlock()
	room, exists := h.Rooms[roomID]
	if !exists {
		log.Printf("Room %s not found for broadcast", roomID)
		return
	}
	for client := range room.Players {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Error broadcasting message to room %s: %v", roomID, err)
			client.Close()
			delete(room.Players, client)
		}
	}

	// Clean up the room if it's empty
	if len(room.Players) == 0 {
		delete(h.Rooms, roomID)
		log.Printf("Room %s deleted because it is empty", roomID)
	}
}
