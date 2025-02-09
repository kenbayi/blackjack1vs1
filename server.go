package main

import (
	"blackjack/config"
	"blackjack/src/db"
	"blackjack/src/handlers"
	"blackjack/src/middlewares"
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	// Initialize PostgreSQL
	db.InitPostgres(cfg.PostgresDSN)

	// Initialize Redis
	db.InitRedis(cfg.RedisAddr, cfg.RedisPass)

	// Setup HTTP router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Protected routes using ValidateSession middleware
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(middlewares.ValidateSession)
	protected.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")
	// Protected GET endpoints
	protected.HandleFunc("/rooms", handlers.HubInstance.GetRooms).Methods("GET")
	protected.HandleFunc("/user/{id}", handlers.DeleteUserHandler).Methods("DELETE")

	// WebSocket route for handling game actions like room creation, joining, etc.
	protected.HandleFunc("/ws", handlers.HandleWebSocket)

	// HTTP server setup
	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Println("HTTP Server is running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown handling for the HTTP server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout for HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP Server shutdown failed: %v", err)
	}

	log.Println("Bye.")
}
