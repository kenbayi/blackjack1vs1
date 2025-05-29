package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"auth_svc/config"
	grpcserver "auth_svc/internal/adapter/grpc/server"
	"auth_svc/internal/adapter/nats/producer"
	postgresrepo "auth_svc/internal/adapter/postgres"
	redisrepo "auth_svc/internal/adapter/redis"
	"auth_svc/internal/usecase"
	natsconn "auth_svc/pkg/nats"
	postgrescon "auth_svc/pkg/postgres"
	redisconn "auth_svc/pkg/redis"
	"auth_svc/pkg/security"
)

const serviceName = "auth-service"

type App struct {
	grpcServer *grpcserver.API
	db         *postgrescon.DB
	natsClient *natsconn.Client
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

	// Initialize NATS producer
	userProducer := producer.NewUserProducer(natsClient, cfg.Nats.NatsSubjects.UserCreatedSubject,
		cfg.Nats.NatsSubjects.UserUpdatedSubject, cfg.Nats.NatsSubjects.UserDeletedSubject,
		cfg.Nats.NatsSubjects.EmailChangeSubject)
	// Initialize Redis
	redisClient, err := redisconn.NewClient(ctx, (redisconn.Config)(cfg.Redis))
	if err != nil {
		return nil, fmt.Errorf("redisconn.NewClient: %w", err)
	}
	log.Println("Redis is connected:", redisClient.Ping(ctx) == nil)

	// Initialize repositories
	userRepo := postgresrepo.NewUserRepository(postgresDB.Conn)
	tokenRepo := postgresrepo.NewRefreshTokenRepository(postgresDB.Conn)
	emailRepo := redisrepo.NewEmail(redisClient, cfg.EmailRedis.ClientTTL)
	// Initialize security components
	jwtManager := security.NewJWTManager(cfg.JWTManager.SecretKey)
	passwordManager := security.NewPasswordManager()

	// Initialize use cases
	userUsecase := usecase.NewUser(
		userRepo,
		tokenRepo,
		emailRepo,
		userProducer,
		transactor.WithinTransaction,
		jwtManager,
		passwordManager,
	)

	// Initialize gRPC server
	gRPCServer := grpcserver.New(
		cfg.Server.GRPCServer,
		userUsecase,
		cfg.JWTManager.SecretKey,
	)

	app := &App{
		grpcServer: gRPCServer,
		db:         postgresDB,
		natsClient: natsClient,
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
