package server

import "statistics/internal/adapter/grpc/server/frontend"

type StatisticUsecase interface {
	frontend.StatisticsUseCase
}
