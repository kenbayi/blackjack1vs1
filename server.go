package main

import (
	"blackjack/config"
	"blackjack/src/db"
	"blackjack/src/handlers"
	"blackjack/src/middlewares"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
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

	// HTTP server setup
	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Running the server in a goroutine
	go func() {
		log.Println("Server is running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Bye.")
}
