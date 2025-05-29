package main

import (
	"blackjack/config"
	"blackjack/src/db"
	"blackjack/src/handlers"
	"blackjack/src/middlewares"
	"context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	db.InitPostgres(cfg.PostgresDSN)
	db.InitRedis(cfg.RedisAddr, cfg.RedisPass)

	router := mux.NewRouter()

	// Enable CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Change to your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Public routes
	router.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Protected routes using ValidateSession middleware
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(middlewares.ValidateSession)
	protected.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")
	protected.HandleFunc("/rooms", handlers.HubInstance.GetRooms).Methods("GET")
	protected.HandleFunc("/history/{id}", handlers.GetHistory).Methods("GET")
	protected.HandleFunc("/user/{username}", handlers.GetUserByUsername).Methods("GET")
	protected.HandleFunc("/updProfile", handlers.UpdateUserProfile).Methods("PUT")
	protected.HandleFunc("/updBalance", handlers.UpdateUserBalance).Methods("PUT")
	protected.HandleFunc("/user/{id}", handlers.DeleteUserHandler).Methods("DELETE")
	protected.HandleFunc("/session", handlers.SessionHandler).Methods("GET")

	protected.HandleFunc("/ws", handlers.HandleWebSocket)

	server := &http.Server{
		Addr:         ":3000",
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("HTTP Server is running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server failed to start: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP Server shutdown failed: %v", err)
	}

	log.Println("Bye.")
}
