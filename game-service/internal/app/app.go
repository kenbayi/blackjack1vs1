package app

import (
	"context"
	"fmt"
	"game_svc/internal/adapter/nats/producer"
	"game_svc/pkg/security"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"game_svc/config"
	usersvc "game_svc/internal/adapter/grpc/server/frontend/proto/user"
	grpcusersvcclient "game_svc/internal/adapter/grpc/users"
	redisrepo "game_svc/internal/adapter/redis"
	wsserver "game_svc/internal/adapter/ws/server"
	"game_svc/internal/usecase"
	grpcconn "game_svc/pkg/grpcconn"
	natsconn "game_svc/pkg/nats"
	redisconn "game_svc/pkg/redis"
	gameservicews "game_svc/pkg/ws"
)

const serviceName = "game-service"

type App struct {
	webSocketServer *wsserver.WebSocketServer
	wsHub           *gameservicews.Hub
	redis           *redisconn.Client
	natsClient      *natsconn.Client
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf("Starting %s application initialization", serviceName)
	log.Println("Initializing JWTManager...")
	if cfg.JWTManager.SecretKey == "" {
		return nil, fmt.Errorf("JWT secret key is not configured")
	}
	jwtManager := security.NewJWTManager(cfg.JWTManager.SecretKey)

	//Initialize gRPC clients
	clientServiceGRPCConn, err := grpcconn.New(cfg.GRPC.GRPCClient.UserServiceURL)
	if err != nil {
		return nil, err
	}
	clientServiceClient := grpcusersvcclient.NewClient(usersvc.NewUserServiceClient(clientServiceGRPCConn))

	// Initialize NATS user
	log.Printf("connecting to NATS hosts: %s\n", strings.Join(cfg.Nats.Hosts, ","))
	natsClient, err := natsconn.NewClient(ctx, cfg.Nats.Hosts, cfg.Nats.NKey, cfg.Nats.IsTest)
	if err != nil {
		return nil, fmt.Errorf("NATS connection failed: %w", err)
	}
	log.Printf("NATS connection status: %s\n", natsClient.Conn.Status().String())

	// Initialize NATS producer
	gameProducer := producer.NewGameEvent(natsClient, cfg.Nats.NatsSubjects.GameResultSubject)

	// 2. Initialize Redis connection
	log.Println("Initializing Redis connection...")
	redisClient, err := redisconn.NewClient(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis NewClient failed: %w", err)
	}
	if errPing := redisClient.Ping(ctx); errPing != nil {
		return nil, fmt.Errorf("redis ping failed: %w", errPing)
	}
	log.Println("Redis connected.")

	// 3. Initialize Repositories
	log.Println("Initializing repositories...")
	roomStateRepo := redisrepo.NewRoomStateRepoImpl(redisClient)
	rankedRepo := redisrepo.NewRankedRepoImpl(redisClient)
	// 4. Initialize Use Cases
	log.Println("Initializing use cases...")
	roomUseCase := usecase.NewRoomService(roomStateRepo, clientServiceClient)               // Ensure NewRoomService matches this
	gameUseCase := usecase.NewGameService(roomStateRepo, gameProducer, clientServiceClient) // Ensure NewGameService matches
	rankedUseCase := usecase.NewRankedUseCase(rankedRepo, clientServiceClient, roomUseCase)
	// 5. Initialize WebSocket Hub
	log.Println("Initializing WebSocket Hub...")
	// The hub itself doesn't directly need messageHandler at construction if it's set later
	// NewHub(messageHandler, onDisconnectHandler)
	hub := gameservicews.NewHub(nil, nil) // Will set handlers next

	// 6. Initialize GameMessageHandler
	log.Println("Initializing GameMessageHandler...")
	gameMessageHandler := wsserver.NewGameMessageHandler(hub, roomUseCase, gameUseCase, rankedUseCase)

	// 7. Set Hub's handlers
	hub.MessageHandler = gameMessageHandler.Handle
	hub.OnDisconnectHandler = func(client *gameservicews.Client) {
		if client.RoomID != "" {
			log.Printf("App: Handling disconnect for UserID: %s, RoomID: %s", client.UserID, client.RoomID)
			go gameMessageHandler.HandlePlayerDisconnect(client.UserID, client.RoomID)
		} else {
			log.Printf("App: Client UserID %s disconnected, was not in a room.", client.UserID)
		}
	}

	// 8. Initialize WebSocket Server
	log.Println("Initializing WebSocket server component...")
	wsServer := wsserver.New(
		cfg.Server, // Pass the ServerConfig part
		hub,
		gameMessageHandler,
		jwtManager,
	)

	log.Printf("%s application initialized successfully.", serviceName)
	return &App{
		webSocketServer: wsServer,
		wsHub:           hub,
		redis:           redisClient,
		natsClient:      natsClient,
	}, nil
}

// Run starts all application components.
func (a *App) Run() error {
	errCh := make(chan error, 1) // Buffered channel for server errors

	// Start the WebSocket Hub's processing loop
	log.Println("Starting WebSocket Hub...")
	go a.wsHub.Run()

	// Start the WebSocket HTTP server
	log.Println("Starting WebSocket server...")
	a.webSocketServer.Run(errCh) // This now runs its ListenAndServe in a goroutine

	log.Printf("Service %s listeners started. Awaiting signals or errors.", serviceName)

	// Wait for errors or shutdown signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("Critical error received: %v. Shutting down.", err)
		// Attempt graceful shutdown even on error
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.webSocketServer.Cfg.ShutdownTimeoutSec)
		defer cancel()
		a.shutdown(shutdownCtx)
		return fmt.Errorf("%s service run failed: %w", serviceName, err)

	case sig := <-signalCh:
		log.Printf("Signal %v received. Initiating graceful shutdown...", sig)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.webSocketServer.Cfg.ShutdownTimeoutSec)
		defer cancel()
		a.shutdown(shutdownCtx)
		log.Println("Graceful shutdown completed.")
	}
	return nil
}

// shutdown attempts to gracefully close all resources.
func (a *App) shutdown(ctx context.Context) {
	log.Println("Executing application shutdown sequence...")

	// Stop WebSocket HTTP server
	if a.webSocketServer != nil {
		if err := a.webSocketServer.Stop(ctx); err != nil { // Stop server
			log.Printf("Error stopping WebSocket server: %v", err)
		}
	}

	// Close Hub (if it has a Close method for resource cleanup, e.g., closing channels)
	//if a.wsHub != nil && hasattr(a.wsHub, "Close") { a.wsHub.Close(); }

	// Close Redis client
	if a.redis != nil {
		log.Println("Closing Redis connection...")
		if err := a.redis.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}

	log.Println("Application shutdown finalized.")
}

// Close is the public method to trigger shutdown.
func (a *App) Close(ctx context.Context) {
	log.Println("App.Close called by external trigger or signal handler.")
	a.shutdown(ctx)
}
