package server

import (
	"api-gateway/internal/adapter/http/server/handler"
	"api-gateway/internal/adapter/http/server/middleware"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/config"
	"github.com/gin-gonic/gin"
)

const serverIPAddress = "0.0.0.0:%d" // Changed to 0.0.0.0 for external access

type API struct {
	server           *gin.Engine
	cfg              config.HTTPServer
	jwt              config.JWTManager
	addr             string
	authHandler      *handler.User
	userHandler      *handler.UserProfile
	statisticHandler *handler.Statistics
}

func New(cfg config.Server, auth UserUsecase, statistic StatisticsUsecase, user UserProfileUsecase) *API {
	// Setting the Gin mode
	gin.SetMode(cfg.HTTPServer.Mode)
	// Creating a new Gin Engine
	server := gin.New()

	// Applying middleware
	server.Use(gin.Recovery())

	// Binding presenter
	authHandler := handler.NewUser(auth)
	userHandler := handler.NewUserProfile(user)
	statisticHandler := handler.NewStatistics(statistic)

	api := &API{
		server:           server,
		cfg:              cfg.HTTPServer,
		jwt:              cfg.JWTManager,
		addr:             fmt.Sprintf(serverIPAddress, cfg.HTTPServer.Port),
		authHandler:      authHandler,
		userHandler:      userHandler,
		statisticHandler: statisticHandler,
	}

	api.setupRoutes()

	return api
}

func (a *API) setupRoutes() {
	// --- Public Routes ---
	// Group for authentication endpoints that do not require a JWT token.
	publicV1 := a.server.Group("/api/v1/auth")
	{
		publicV1.POST("/register", a.authHandler.Register)
		publicV1.POST("/login", a.authHandler.Login)
		publicV1.POST("/refresh", a.authHandler.RefreshToken)
		publicV1.POST("/password-reset/request", a.authHandler.RequestPasswordReset)
		publicV1.POST("/password-reset/confirm", a.authHandler.ResetPassword)
		publicV1.POST("/email-change/confirm", a.authHandler.ConfirmEmailChange)
	}

	// --- Private Routes ---
	// Group for all endpoints that require JWT authentication.
	v1 := a.server.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(a.jwt.SecretKey))
	{
		// These are handled by authHandler (*handler.User).
		userMeGroup := v1.Group("/user/me")
		{
			userMeGroup.DELETE("/delete", a.authHandler.DeleteByID)
			userMeGroup.PATCH("/username", a.authHandler.UpdateUsername)
			userMeGroup.POST("/password", a.authHandler.ChangePassword)
			userMeGroup.POST("/email-change/request", a.authHandler.UpdateEmailRequest)
		}

		// These are handled by userHandler (*handler.UserProfile).
		usersGroup := v1.Group("/users")
		{
			usersGroup.GET("/profile", a.userHandler.GetProfile)
			usersGroup.PATCH("/profile", a.userHandler.UpdateProfile)
			usersGroup.GET("/balance", a.userHandler.GetBalance)
			usersGroup.GET("/rating", a.userHandler.GetRating)
		}

		// Routes for game statistics
		// These are handled by statisticHandler (*handler.Statistics).
		statisticsGroup := v1.Group("/statistics")
		{
			statisticsGroup.GET("/general", a.statisticHandler.GetGeneralGameStats)
			statisticsGroup.GET("/user/:userID", a.statisticHandler.GetUserGameStats)
			statisticsGroup.GET("/leaderboard", a.statisticHandler.GetLeaderboard)
		}
	}
}

func (a *API) Run(errCh chan<- error) {
	go func() {
		log.Printf("HTTP server starting on: %v", a.addr)

		// No need to reinitialize `a.server` here. Just run it directly.
		if err := a.server.Run(a.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start HTTP server: %w", err)
			return
		}
	}()
}

func (a *API) Stop() error {
	// Setting up the signal channel to catch termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocking until a signal is received
	sig := <-quit
	log.Println("Shutdown signal received", "signal:", sig.String())

	// Creating a context with timeout for graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("HTTP server shutting down gracefully")

	// Note: You can use `Shutdown` if you use `http.Server` instead of `gin.Engine`.
	log.Println("HTTP server stopped successfully")

	return nil
}
