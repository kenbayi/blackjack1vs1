package ws

import (
	"log"
	"sync"
)

// MessageHandlerFunc - это тип функции, которая будет обрабатывать входящие RawMessage.
// Это позволяет отделить логику хаба от специфической логики обработки сообщений (например, игровой).
type MessageHandlerFunc func(msg *RawMessage)
type OnDisconnectHandlerFunc func(client *Client)

// Hub управляет набором активных клиентов и рассылает им сообщения.
// Этот Hub является общим и не знает о специфических игровых комнатах или логике игры.
// Управление комнатами и игровая логика будут обрабатываться вышестоящими сервисами (use cases),
// которые получат сообщения через MessageHandler.
type Hub struct {
	// Зарегистрированные клиенты. Ключ - указатель на Client, значение - bool (для использования как set).
	clients         map[*Client]bool
	clientsByUserID map[string]*Client
	// Канал для входящих "сырых" сообщений от клиентов.
	// Эти сообщения будут переданы в MessageHandler.
	Broadcast chan *RawMessage

	// Канал для регистрации новых клиентов.
	Register chan *Client

	// Канал для отмены регистрации клиентов.
	Unregister chan *Client

	// Mutex для защиты доступа к карте clients.
	mu sync.Mutex

	// messageHandler - функция, которая будет вызвана для обработки сообщений.
	MessageHandler MessageHandlerFunc

	OnDisconnectHandler OnDisconnectHandlerFunc
}

func (h *Hub) GetClientByUserID(userID string) (*Client, bool) {
	client, ok := h.clientsByUserID[userID]
	return client, ok
}

// NewHub создает новый экземпляр Hub.
// messageHandler - это функция, которая будет обрабатывать каждое входящее сообщение.
func NewHub(handler MessageHandlerFunc, onDisconnectHandler OnDisconnectHandlerFunc) *Hub {
	if handler == nil {
		// Предоставляем обработчик по умолчанию, если не указан, чтобы избежать паники nil pointer.
		// Этот обработчик просто логирует получение сообщения.
		handler = func(msg *RawMessage) {
			log.Printf("Hub received message from client %s, but no specific handler is set. Payload: %s", msg.Client.UserID, string(msg.Payload))
		}
	}
	return &Hub{
		Broadcast:           make(chan *RawMessage, 256), // Буферизированный канал
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		clients:             make(map[*Client]bool),
		clientsByUserID:     make(map[string]*Client),
		MessageHandler:      handler,
		OnDisconnectHandler: onDisconnectHandler,
	}
}

// Run запускает главный цикл Hub для обработки входящих событий.
func (h *Hub) Run() {
	log.Println("WebSocket Hub is running...")
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.clientsByUserID[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Hub: Client registered: UserID %s, RemoteAddr: %s", client.UserID, client.Conn.RemoteAddr().String())

		case client := <-h.Unregister: // Клиент отключается (либо сам, либо из-за ошибки в ReadPump/WritePump)
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientsByUserID, client.UserID)
				close(client.Send) // Важно закрыть канал, чтобы WritePump завершился
				log.Printf("Hub: Client unregistered: UserID %s, RoomID: %s, RemoteAddr: %s", client.UserID, client.RoomID, client.Conn.RemoteAddr().String())

				// Вызываем коллбэк OnDisconnectHandler, если он установлен
				if h.OnDisconnectHandler != nil { // <<<< НОВАЯ ЛОГИКА
					// Запускаем в горутине, чтобы не блокировать цикл хаба,
					// если обработчик дисконнекта долгий.
					go h.OnDisconnectHandler(client)
				}
			}
			h.mu.Unlock()

		case rawMsg := <-h.Broadcast:
			if h.MessageHandler != nil {
				// Обработка сообщения в новой горутине, чтобы не блокировать хаб
				go h.MessageHandler(rawMsg)
			} else {
				log.Printf("Hub: No message handler configured for message from client %s.", rawMsg.Client.UserID)
			}
		}
	}
}

// BroadcastToAll отправляет сообщение всем подключенным клиентам.
func (h *Hub) BroadcastToAll(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for client := range h.clients {
		select {
		case client.Send <- message:
			count++
		default: // Если буфер отправки клиента заполнен, пропускаем или удаляем клиента
			log.Printf("Client %s (UserID: %s) send buffer full or closed, removing client during BroadcastToAll.", client.Conn.RemoteAddr(), client.UserID)
			close(client.Send)
			delete(h.clients, client)
		}
	}
	log.Printf("BroadcastToAll: Message sent to %d clients.", count)
}

// BroadcastToRoom отправляет сообщение всем клиентам в указанной комнате.
// Требует, чтобы у Client был установлен RoomID.
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	if roomID == "" {
		log.Println("BroadcastToRoom: Attempted to broadcast to empty roomID. Message not sent.")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	log.Printf("Attempting to broadcast to room '%s', message: %s", roomID, string(message))
	for client := range h.clients {
		if client.RoomID == roomID {
			select {
			case client.Send <- message:
				count++
				log.Printf("Sent message to client %s (UserID: %s) in room %s", client.Conn.RemoteAddr(), client.UserID, roomID)
			default:
				log.Printf("Client %s (UserID: %s) in room %s send buffer full or closed, removing client.", client.Conn.RemoteAddr(), client.UserID, roomID)
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	if count > 0 {
		log.Printf("BroadcastToRoom: Message sent to %d clients in room %s.", count, roomID)
	} else {
		log.Printf("BroadcastToRoom: No clients found in room %s or message not sent.", roomID)
	}
}

// BroadcastToClient отправляет сообщение конкретному клиенту.
func (h *Hub) BroadcastToClient(targetClient *Client, message []byte) {
	if targetClient == nil {
		log.Println("BroadcastToClient: targetClient is nil.")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверяем, что клиент все еще зарегистрирован
	if _, ok := h.clients[targetClient]; ok {
		select {
		case targetClient.Send <- message:
			log.Printf("Sent direct message to client %s (UserID: %s)", targetClient.Conn.RemoteAddr(), targetClient.UserID)
		default:
			log.Printf("Target client %s (UserID: %s) send buffer full or closed, removing client.", targetClient.Conn.RemoteAddr(), targetClient.UserID)
			close(targetClient.Send)
			delete(h.clients, targetClient)
		}
	} else {
		log.Printf("Target client %s (UserID: %s) not found or already unregistered for direct message.", targetClient.Conn.RemoteAddr(), targetClient.UserID)
	}
}

// GetClientCount возвращает количество подключенных клиентов.
func (h *Hub) GetClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}
