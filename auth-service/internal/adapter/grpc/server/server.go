package server

import (
	frontendsvc "auth_svc/internal/adapter/grpc/server/frontend/proto/user"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"auth_svc/config"
	"auth_svc/internal/adapter/grpc/server/frontend"
)

type API struct {
	s               *grpc.Server
	cfg             config.GRPCServer
	secretKey       string
	addr            string
	customerUsecase CustomerUsecase
}

func New(
	cfg config.GRPCServer,
	customerUsecase CustomerUsecase,
	secretKey string,
) *API {
	return &API{
		cfg:             cfg,
		addr:            fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		customerUsecase: customerUsecase,
		secretKey:       secretKey,
	}
}

func (a *API) Run(ctx context.Context, errCh chan<- error) {
	go func() {
		log.Println(ctx, "gRPC server starting listen", fmt.Sprintf("addr: %s", a.addr))

		if err := a.run(ctx); err != nil {
			errCh <- fmt.Errorf("can't start grpc server: %w", err)

			return
		}
	}()
}

// Stop method gracefully stops grpc API server. Provide context to force stop on timeout.
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
	case <-ctx.Done(): // Stop immediately if the context is terminated
		a.s.Stop()
	case <-stopped:
	}

	return nil
}

// run starts and runs GRPCServer server.
func (a *API) run(ctx context.Context) error {
	a.s = grpc.NewServer(a.setOptions(ctx)...)

	// Register bo services
	frontendsvc.RegisterUserServiceServer(a.s, frontend.NewUser(a.customerUsecase))

	// Register reflection service
	reflection.Register(a.s)

	listener, err := net.Listen("tcp", a.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	err = a.s.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}
