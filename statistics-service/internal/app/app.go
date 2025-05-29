package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"statistics/config"
	grpcserver "statistics/internal/adapter/grpc"
	mongorepo "statistics/internal/adapter/mongo"
	natshandler "statistics/internal/adapter/nats/handler"
	"statistics/internal/adapter/redis"
	"statistics/internal/usecase"
	mongocon "statistics/pkg/mongo"
	natsconn "statistics/pkg/nats"
	natsconsumer "statistics/pkg/nats/consumer"
	redisconn "statistics/pkg/redis"
)

const serviceName = "statistics-service"

type App struct {
	grpcServer         *grpcserver.API
	natsPubSubConsumer *natsconsumer.PubSub
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Println(fmt.Sprintf("starting %v service", serviceName))

	log.Println("connecting to mongo", "database", cfg.Mongo.Database)
	mongoDB, err := mongocon.NewDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("mongo: %w", err)
	}

	// nats user
	log.Println("connecting to NATS", "hosts", strings.Join(cfg.Nats.Hosts, ","))
	natsClient, err := natsconn.NewClient(ctx, cfg.Nats.Hosts, cfg.Nats.NKey, cfg.Nats.IsTest)
	if err != nil {
		return nil, fmt.Errorf("nats.NewClient: %w", err)
	}
	log.Println("NATS connection status is", natsClient.Conn.Status().String())

	// redis user
	redisClient, err := redisconn.NewClient(ctx, (redisconn.Config)(cfg.Redis))
	if err != nil {
		return nil, fmt.Errorf("redisconn.NewClient: %w", err)
	}
	log.Println("Redis is connected:", redisClient.Ping(ctx) == nil)
	// Setup NATS subscriber for order events

	natsPubSubConsumer := natsconsumer.NewPubSub(natsClient)

	// redis cache
	statisticsRedisCache := redis.NewStatisticsRedisCache(redisClient)

	// Initialize statistics components
	statsRepo := mongorepo.NewStatisticsRepository(mongoDB.Conn)
	historyStatsRepo := mongorepo.NewGameHistoryRepository(mongoDB.Conn)
	statsUsecase := usecase.NewStatisticsUseCase(statsRepo, statisticsRedisCache, historyStatsRepo)
	userHandler := natshandler.NewEventHandler(statsUsecase)

	subscriptions := []natsconsumer.PubSubSubscriptionConfig{
		{
			Subject: "user.events.created",
			Handler: userHandler.HandleNATSUserCreated,
		},
		{
			Subject: "game.events.result",
			Handler: userHandler.HandleNATSGameResult,
		},
		{
			Subject: "user.events.deleted",
			Handler: userHandler.HandleNATSUserDeleted,
		},
	}

	for _, sub := range subscriptions {
		natsPubSubConsumer.Subscribe(sub)
	}

	gRPCServer := grpcserver.New(
		cfg.Server.GRPCServer,
		statsUsecase,
	)

	app := &App{
		grpcServer:         gRPCServer,
		natsPubSubConsumer: natsPubSubConsumer,
	}

	return app, nil
}

func (a *App) Close(ctx context.Context) {
	err := a.grpcServer.Stop(ctx)
	if err != nil {
		log.Println("failed to shutdown gRPC service", err)
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()
	a.grpcServer.Run(ctx, errCh)
	a.natsPubSubConsumer.Start(ctx, errCh)
	log.Println(fmt.Sprintf("service %v started", serviceName))
	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun

	case s := <-shutdownCh:
		log.Println(fmt.Sprintf("received signal: %v. Running graceful shutdown...", s))

		a.Close(ctx)
		log.Println("graceful shutdown completed!")
	}

	return nil
}
