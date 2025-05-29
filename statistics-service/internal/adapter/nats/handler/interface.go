package handler

import (
	"context"
	"statistics/internal/model"
)

type StatisticsEventConsumer interface {
	HandleUserCreated(ctx context.Context, eventData model.UserCreatedEventData) error
	HandleUserDeleted(ctx context.Context, eventData model.UserDeletedEventData) error
	HandleGameResult(ctx context.Context, eventData model.GameResultEventData) error
}
