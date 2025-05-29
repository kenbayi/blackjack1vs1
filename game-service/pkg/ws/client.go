package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client представляет собой подключенного WebSocket клиента.
type Client struct {
	Hub *Hub

	// WebSocket соединение.
	Conn *websocket.Conn

	// Буферизированный канал исходящих сообщений.
	Send chan []byte

	// UserID пользователя, связанного с этим клиентом.
	UserID string

	// RoomID комнаты, в которой находится клиент.
	RoomID string
}

// ReadPump считывает сообщения от WebSocket соединения и передает их в хаб.
// Запускается в отдельной горутине для каждого соединения.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing connection in ReadPump for client %s: %v", c.UserID, err)
		}
		log.Printf("ReadPump stopped for client %s (UserID: %s)", c.Conn.RemoteAddr(), c.UserID)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("Error setting read deadline for client %s: %v", c.UserID, err)
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("Error setting read deadline in pong handler for client %s: %v", c.UserID, err)
			// It might be better to return the error here to stop the ReadPump
			// if the deadline cannot be extended.
		}
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error for client %s (UserID: %s): %v", c.Conn.RemoteAddr(), c.UserID, err)
			} else {
				log.Printf("Read error for client %s (UserID: %s): %v", c.Conn.RemoteAddr(), c.UserID, err)
			}
			break // Выход из цикла при ошибке чтения или закрытии соединения
		}

		// Создаем RawMessage для передачи в хаб
		// Игровая логика (парсер JSON и т.д.) будет вызываться обработчиком в хабе
		rawMessage := &RawMessage{
			Client:  c,
			Payload: message,
		}
		// Отправляем сообщение в канал Broadcast хаба
		// Важно: убедитесь, что канал Broadcast хаба читается, иначе здесь будет блокировка
		select {
		case c.Hub.Broadcast <- rawMessage:
		default:
			log.Printf("Hub broadcast channel full, dropping message from client %s", c.UserID)
			// Можно рассмотреть вариант закрытия соединения, если хаб не справляется
		}
	}
}

// WritePump отправляет сообщения из хаба в WebSocket соединение.
// Запускается в отдельной горутине для каждого соединения.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing connection in WritePump for client %s: %v", c.UserID, err)
		}
		log.Printf("WritePump stopped for client %s (UserID: %s)", c.Conn.RemoteAddr(), c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("Error setting write deadline for client %s: %v", c.UserID, err)
				return // Выход при ошибке установки deadline
			}
			if !ok {
				// Канал Send был закрыт хабом.
				log.Printf("Client %s (UserID: %s) send channel closed.", c.Conn.RemoteAddr(), c.UserID)
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("Error writing close message for client %s: %v", c.UserID, err)
				}
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message to client %s (UserID: %s): %v", c.Conn.RemoteAddr(), c.UserID, err)
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("Error setting write deadline for ping for client %s: %v", c.UserID, err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s (UserID: %s): %v", c.Conn.RemoteAddr(), c.UserID, err)
				return // Выход при ошибке отправки ping
			}
		}
	}
}
