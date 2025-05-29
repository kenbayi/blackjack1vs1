package server

import (
	"context"
	"errors"
	"fmt"
	"game_svc/pkg/security"
	"log"
	"net/http"

	"game_svc/config"
	gameservicews "game_svc/pkg/ws"
	"github.com/gorilla/websocket"
)

const defaultWebSocketPath = "/ws"

// WebSocketServer manages the HTTP server for WebSocket connections.
type WebSocketServer struct {
	httpServer *http.Server
	hub        *gameservicews.Hub
	handler    *GameMessageHandler
	Cfg        config.ServerConfig
	wsPath     string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking using cfg.AllowedOrigins
		log.Printf("WebSocket CheckOrigin: Host %s, Origin %s", r.Host, r.Header.Get("Origin"))
		return true
	},
}

// New creates a new instance of WebSocketServer.
func New(
	cfg config.ServerConfig,
	hub *gameservicews.Hub,
	gameHandler *GameMessageHandler,
	jwtManager *security.JWTManager,
) *WebSocketServer {
	mux := http.NewServeMux()

	wsPath := cfg.WebSocketPath
	if wsPath == "" {
		wsPath = defaultWebSocketPath
	}

	// Оборачиваем наш основной обработчик WebSocket в AuthJWTMiddleware
	wsHandlerWithAuth := AuthJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWsLogic(hub, gameHandler, w, r)
	}), jwtManager) // <<<< Передаем jwtManager в middleware

	mux.Handle(wsPath, wsHandlerWithAuth)

	addr := ":" + cfg.WebSocketPort
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeoutSec,
		WriteTimeout: cfg.WriteTimeoutSec,
		IdleTimeout:  cfg.IdleTimeoutSec,
	}
	log.Printf("WebSocketServer configured to listen on %s%s with JWT Auth", addr, wsPath)
	return &WebSocketServer{
		httpServer: httpSrv,
		hub:        hub,
		handler:    gameHandler,
		Cfg:        cfg,
		wsPath:     wsPath,
	}
}

// serveWsLogic handles the upgrade of an HTTP connection to a WebSocket connection.
func serveWsLogic(hub *gameservicews.Hub, gameHandler *GameMessageHandler, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error from %s: %v", r.RemoteAddr, err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	userIDFromCtx := r.Context().Value(UserIDKey)

	userID, ok := userIDFromCtx.(string)
	if !ok || userID == "" {
		log.Printf("serveWsLogic: UserID not found in context or is not a string for %s. This indicates an issue with the auth middleware or context propagation.", r.RemoteAddr)
		http.Error(w, "Internal Server Error: User identification missing after authentication phase.", http.StatusInternalServerError)
		return
	}

	client := &gameservicews.Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256), // Buffered channel
		UserID: userID,
		// RoomID will be set by game logic via messages
	}

	log.Printf("Client connected: UserID %s, RemoteAddr: %s", client.UserID, client.Conn.RemoteAddr().String())
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
	// Disconnect logic is now handled via hub.OnDisconnectHandler set in app.go
}

// Run starts the WebSocket HTTP server.
func (s *WebSocketServer) Run(errCh chan<- error) {
	go func() {
		log.Printf("WebSocket HTTP server starting on address: %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("WebSocket HTTP server ListenAndServe failed: %w", err)
			return
		}
		log.Println("WebSocket HTTP server has been stopped.")
	}()
}

// Stop gracefully shuts down the WebSocket HTTP server.
func (s *WebSocketServer) Stop(ctx context.Context) error {
	log.Println("Attempting to shut down WebSocket HTTP server gracefully...")

	// Use the timeout from config for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, s.Cfg.ShutdownTimeoutSec)
	defer cancel()

	err := s.httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("WebSocket HTTP server graceful shutdown failed: %v", err)
		return fmt.Errorf("WebSocket HTTP server shutdown error: %w", err)
	}
	log.Println("WebSocket HTTP server stopped gracefully.")
	return nil
}

// GetAddr returns the address the server is configured to listen on.
func (s *WebSocketServer) GetAddr() string {
	return s.httpServer.Addr
}
