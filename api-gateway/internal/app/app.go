package app

import (
	"api-gateway/internal/usecase"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"api-gateway/config"
	authsvc "api-gateway/internal/adapter/frontend/proto/auth"
	statisticsvc "api-gateway/internal/adapter/frontend/proto/statistics"
	usersvc "api-gateway/internal/adapter/frontend/proto/user"
	grpcauthsvcclient "api-gateway/internal/adapter/grpc/auth"
	grpcstatisticssvcclient "api-gateway/internal/adapter/grpc/statistics"
	grpcusersvcclient "api-gateway/internal/adapter/grpc/user"
	httpserver "api-gateway/internal/adapter/http/server"
	"api-gateway/pkg/grpcconn"
)

const serviceName = "api-gateway"

type App struct {
	httpServer *httpserver.API
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log.Println(fmt.Sprintf("starting %v service", serviceName))

	authServiceGRPCConn, err := grpcconn.New(cfg.GRPC.GRPCClient.AuthServiceURL)
	if err != nil {
		return nil, err
	}

	userServiceGRPCConn, err := grpcconn.New(cfg.GRPC.GRPCClient.UserServiceURL)
	if err != nil {
		return nil, err
	}

	statsServiceGRPCConn, err := grpcconn.New(cfg.GRPC.GRPCClient.StatisticsServiceURL)
	if err != nil {
		return nil, err
	}

	authServiceClient := grpcauthsvcclient.NewAuth(authsvc.NewUserServiceClient(authServiceGRPCConn))
	userServiceClient := grpcusersvcclient.NewUser(usersvc.NewUserServiceClient(userServiceGRPCConn))
	statisticServiceClient := grpcstatisticssvcclient.NewStatistics(
		statisticsvc.NewStatisticsServiceClient(statsServiceGRPCConn),
	)

	authUsecase := usecase.NewUser(authServiceClient)
	inventoryUsecase := usecase.NewUserProfile(userServiceClient)
	statisticUsecase := usecase.NewStatistics(statisticServiceClient)

	// http service
	httpServer := httpserver.New(cfg.Server, authUsecase, statisticUsecase, inventoryUsecase)

	app := &App{
		httpServer: httpServer,
	}

	return app, nil
}

func (a *App) Close(ctx context.Context) {
	err := a.httpServer.Stop()
	if err != nil {
		log.Println("failed to shutdown gRPC service", err)
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()
	a.httpServer.Run(errCh)

	log.Println(fmt.Sprintf("service %v started", serviceName))

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

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
