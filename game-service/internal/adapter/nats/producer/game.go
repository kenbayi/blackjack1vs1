package producer

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"time"

	"google.golang.org/protobuf/proto"

	"game_svc/internal/adapter/nats/producer/dto"
	"game_svc/internal/model"
	"game_svc/pkg/nats"
)

const PushTimeout = time.Second * 30

type GameEvent struct {
	natsClient        *nats.Client
	gameResultSubject string
}

func NewGameEvent(
	natsClient *nats.Client,
	gameResultSubject string,
) *GameEvent {
	return &GameEvent{
		natsClient:        natsClient,
		gameResultSubject: gameResultSubject,
	}
}

func (c *GameEvent) PushGameEnd(ctx context.Context, results *model.Result, bet int64) error {
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	pbEvent := dto.FromResult(results, bet)
	// --- Логирование Protobuf объекта (в читаемом JSON формате) ---
	jsonMarshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	jsonBytes, errJson := jsonMarshaler.Marshal(pbEvent)
	if errJson != nil {
		log.Printf("Error marshalling pbEvent to JSON for logging: %v", errJson)
	} else {
		log.Printf("PushGameEnd: Publishing Protobuf event (as JSON for readability):\n%s", string(jsonBytes))
	}
	// --- Конец логирования Protobuf объекта ---

	data, err := proto.Marshal(pbEvent)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	err = c.natsClient.Conn.Publish(c.gameResultSubject, data)
	if err != nil {
		return fmt.Errorf("publish GameResult: %w", err)
	}
	log.Println("GameResult event pushed")
	return nil
}
