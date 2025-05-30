package handler

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
	"statistics/internal/adapter/nats/handler/dto"
)

type EventHandler struct {
	statsUsecase StatisticsEventConsumer
}

func NewEventHandler(uc StatisticsEventConsumer) *EventHandler {
	return &EventHandler{statsUsecase: uc}
}

// HandleNATSUserCreated processes UserCreated events from NATS.
func (h *EventHandler) HandleNATSUserCreated(ctx context.Context, msg *nats.Msg) error {
	log.Printf("NATS Handler: Received UserCreated event, Subject: %s", msg.Subject)

	eventData, err := dto.ToUserCreatedEventData(msg.Data)
	if err != nil {
		log.Printf("NATS Handler: Failed to map UserCreated event data: %v. Msg Data: %s", err, string(msg.Data))
		return err
	}

	if err := h.statsUsecase.HandleUserCreated(ctx, *eventData); err != nil {
		log.Printf("NATS Handler: Failed to process UserCreated event in use case for UserID %d: %v", eventData.ID, err)
		return err
	}

	log.Printf("NATS Handler: UserCreated event processed successfully for UserID: %d", eventData.ID)
	return nil
}

// HandleNATSUserDeleted processes UserDeleted events from NATS.
func (h *EventHandler) HandleNATSUserDeleted(ctx context.Context, msg *nats.Msg) error {
	log.Printf("NATS Handler: Received UserDeleted event, Subject: %s", msg.Subject)

	eventData, err := dto.ToUserDeletedEventData(msg.Data)
	if err != nil {
		log.Printf("NATS Handler: Failed to map UserDeleted event data: %v. Msg Data: %s", err, string(msg.Data))
		return err
	}

	if err := h.statsUsecase.HandleUserDeleted(ctx, *eventData); err != nil {
		log.Printf("NATS Handler: Failed to process UserDeleted event in use case for UserID %d: %v", eventData.ID, err)
		return err
	}

	log.Printf("NATS Handler: UserDeleted event processed successfully for UserID: %d", eventData.ID)
	return nil
}

// HandleNATSGameResult processes GameResult events from NATS.
func (h *EventHandler) HandleNATSGameResult(ctx context.Context, msg *nats.Msg) error {
	log.Printf("NATS Handler: Received GameResult event, Subject: %s", msg.Subject)

	eventData, err := dto.ToGameResultEventData(msg.Data)
	if err != nil {
		log.Printf("NATS Handler: Failed to map GameResult event data: %v. Msg Data: %s", err, string(msg.Data))
		return err
	}

	if err := h.statsUsecase.HandleGameResult(ctx, *eventData); err != nil {
		log.Printf("NATS Handler: Failed to process GameResult event in use case for RoomID %s: %v", eventData.RoomID, err)
		return err
	}

	log.Printf("NATS Handler: GameResult event processed successfully for RoomID: %s", eventData.RoomID)
	return nil
}
