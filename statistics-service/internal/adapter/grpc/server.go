package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"statistics/internal/adapter/grpc/server/frontend"
	statisticsproto "statistics/internal/adapter/grpc/server/frontend/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"statistics/config"
)

type API struct {
	s            *grpc.Server
	cfg          config.GRPCServer
	addr         string
	statsUsecase StatisticUsecase
}

func New(
	cfg config.GRPCServer,
	statsUsecase StatisticUsecase,
) *API {
	return &API{
		cfg:          cfg,
		addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		statsUsecase: statsUsecase,
	}
}

func (a *API) Run(ctx context.Context, errCh chan<- error) {
	go func() {
		log.Println("gRPC server starting listen", fmt.Sprintf("addr: %s", a.addr))

		if err := a.run(ctx); err != nil {
			errCh <- fmt.Errorf("can't start grpc server: %w", err)
			return
		}
	}()
}

func (a *API) Stop(ctx context.Context) error {
	if a.s == nil {
		return nil
	}

	stopped := make(chan struct{})
	go func() {
		a.s.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		a.s.Stop()
	case <-stopped:
	}

	return nil
}

func (a *API) run(ctx context.Context) error {
	a.s = grpc.NewServer(a.setOptions(ctx)...)

	// Register services
	statisticsproto.RegisterStatisticsServiceServer(a.s, frontend.NewStatisticsServer(a.statsUsecase))

	reflection.Register(a.s)

	listener, err := net.Listen("tcp", a.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	return a.s.Serve(listener)
}
