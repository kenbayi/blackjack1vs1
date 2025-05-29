package ws

// RawMessage представляет "сырое" сообщение, полученное от клиента.
// Это сообщение будет передано в MessageHandler хаба для дальнейшей обработки
// (например, JSON анмаршалинг и вызов соответствующей игровой логики).
type RawMessage struct {
	Client  *Client // Клиент, отправивший сообщение
	Payload []byte  // "Сырые" байты сообщения
}

// OutboundMessage представляет структурированное сообщение, отправляемое сервером клиенту.
// Это пример, вы можете определить свои структуры для исходящих сообщений.
type OutboundMessage struct {
	Type    string      `json:"type"`    // Тип сообщения (например, "game_update", "error", "player_joined")
	Content interface{} `json:"content"` // Содержимое сообщения
}
