package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"user_svc/config"
	grpcserver "user_svc/internal/adapter/grpc/server"
	natssubscriber "user_svc/internal/adapter/nats/subscriber"
	postgresrepo "user_svc/internal/adapter/postgres"
	redisrepo "user_svc/internal/adapter/redis"
	"user_svc/internal/usecase"
	natsconn "user_svc/pkg/nats"
	natsconsumer "user_svc/pkg/nats/consumer"
	postgrescon "user_svc/pkg/postgres"
	redisconn "user_svc/pkg/redis"
)

const serviceName = "user-service"

type App struct {
	grpcServer         *grpcserver.API
	db                 *postgrescon.DB
	natsPubSubConsumer *natsconsumer.PubSub
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Printf("starting %s service\n", serviceName)

	// Initialize PostgreSQL connection
	log.Printf("connecting to postgres database: %s\n", cfg.Postgres.Database)
	postgresDB, err := postgrescon.NewDB(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}

	// Initialize transactor for transaction handling
	transactor := postgrescon.NewTransactor(postgresDB.Conn)

	// Initialize NATS user
	log.Printf("connecting to NATS hosts: %s\n", strings.Join(cfg.Nats.Hosts, ","))
	natsClient, err := natsconn.NewClient(ctx, cfg.Nats.Hosts, cfg.Nats.NKey, cfg.Nats.IsTest)
	if err != nil {
		return nil, fmt.Errorf("NATS connection failed: %w", err)
	}
	log.Printf("NATS connection status: %s\n", natsClient.Conn.Status().String())

	natsPubSubConsumer := natsconsumer.NewPubSub(natsClient)
	// Initialize Redis
	redisClient, err := redisconn.NewClient(ctx, (redisconn.Config)(cfg.Redis))
	if err != nil {
		return nil, fmt.Errorf("redisconn.NewClient: %w", err)
	}
	log.Println("Redis is connected:", redisClient.Ping(ctx) == nil)

	// Initialize repositories
	userRepo := postgresrepo.NewUserRepository(postgresDB.Conn)
	userCache := redisrepo.NewUserCache(redisClient, cfg.Cache.ClientTTL)
	// Initialize use cases
	userUsecase := usecase.NewUser(
		userRepo,
		transactor.WithinTransaction,
		userCache,
	)
	userHandler := natssubscriber.NewUserSubscriber(userUsecase)

	subscriptions := []natsconsumer.PubSubSubscriptionConfig{
		{
			Subject: "user.events.created",
			Handler: userHandler.UserCreatedEvent,
		},
		{
			Subject: "user.events.updated",
			Handler: userHandler.UserUpdatedEvent,
		},
		{
			Subject: "user.events.deleted",
			Handler: userHandler.UserDeletedEvent,
		},
	}

	for _, sub := range subscriptions {
		natsPubSubConsumer.Subscribe(sub)
	}

	// Initialize gRPC server
	gRPCServer := grpcserver.New(
		cfg.Server.GRPCServer,
		userUsecase,
	)

	app := &App{
		grpcServer:         gRPCServer,
		db:                 postgresDB,
		natsPubSubConsumer: natsPubSubConsumer,
	}

	return app, nil
}

func (a *App) Close(ctx context.Context) {
	// Close gRPC server
	if err := a.grpcServer.Stop(ctx); err != nil {
		log.Printf("failed to shutdown gRPC service: %v\n", err)
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()

	// Start gRPC server
	go a.grpcServer.Run(ctx, errCh)
	go a.natsPubSubConsumer.Start(ctx, errCh)

	log.Printf("service %s started successfully\n", serviceName)

	// Set up signal handling
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("service error: %w", err)

	case sig := <-shutdownCh:
		log.Printf("received signal %v, initiating graceful shutdown...\n", sig)
		a.Close(ctx)
		log.Println("graceful shutdown completed")
	}

	return nil
}
