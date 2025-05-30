package dto

import (
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
	eventsproto "statistics/internal/adapter/grpc/server/frontend/proto/events"
	"statistics/internal/model"
)

// ToUserCreatedEventData maps from a NATS message payload to model.UserCreatedEventData.
func ToUserCreatedEventData(msgData []byte) (*model.UserCreatedEventData, error) {
	var protoEvent eventsproto.UserCreated
	if err := proto.Unmarshal(msgData, &protoEvent); err != nil {
		return nil, fmt.Errorf("proto unmarshal UserCreated event error: %w", err)
	}

	if protoEvent.User == nil {
		return nil, fmt.Errorf("missing user data in UserCreated event")
	}

	// Map from protoEvent.User to model.UserCreatedEventData
	domainEventData := &model.UserCreatedEventData{
		ID:        protoEvent.User.Id,
		Email:     protoEvent.User.Email,
		Username:  protoEvent.User.Username,
		CreatedAt: protoEvent.User.CreatedAt.AsTime(),
		UpdatedAt: protoEvent.User.UpdatedAt.AsTime(),
		IsDeleted: protoEvent.User.IsDeleted,
	}

	return domainEventData, nil
}

// ToUserDeletedEventData maps from a NATS message payload to model.UserDeletedEventData.
func ToUserDeletedEventData(msgData []byte) (*model.UserDeletedEventData, error) {
	var protoEvent eventsproto.UserDeleted // Assuming this is your Protobuf message type for user deletion
	if err := proto.Unmarshal(msgData, &protoEvent); err != nil {
		return nil, fmt.Errorf("proto unmarshal UserDeleted event error: %w", err)
	}

	// Map from protoEvent to model.UserDeletedEventData
	domainEventData := &model.UserDeletedEventData{
		ID: protoEvent.Id, // Assuming proto UserDeleted has Id
	}

	return domainEventData, nil
}

// ToGameResultEventData maps from a NATS message payload to model.GameResultEventData.
func ToGameResultEventData(msgData []byte) (*model.GameResultEventData, error) {
	var protoEvent eventsproto.GameResult // Using GameResult as defined in your example
	if err := proto.Unmarshal(msgData, &protoEvent); err != nil {
		return nil, fmt.Errorf("proto unmarshal GameResult event error: %w", err)
	}

	// Map from protoEvent to model.GameResultEventData
	p1Data := model.PlayerGameResultData{}
	if protoEvent.Player1 != nil {
		p1Data = model.PlayerGameResultData{
			PlayerID:   protoEvent.Player1.PlayerId,
			FinalScore: protoEvent.Player1.FinalScore,
			FinalHand:  protoEvent.Player1.FinalHand, // Already []string in proto
		}
	} else {
		log.Println("Warning: GameResult event received with nil Player1 data")
	}

	p2Data := model.PlayerGameResultData{}
	if protoEvent.Player2 != nil {
		p2Data = model.PlayerGameResultData{
			PlayerID:   protoEvent.Player2.PlayerId,
			FinalScore: protoEvent.Player2.FinalScore,
			FinalHand:  protoEvent.Player2.FinalHand, // Already []string in proto
		}
	} else {
		log.Println("Warning: GameResult event received with nil Player2 data")
	}

	domainEventData := &model.GameResultEventData{
		RoomID:    protoEvent.RoomId,
		WinnerID:  protoEvent.WinnerId,
		LoserID:   protoEvent.LoserId,
		Bet:       protoEvent.Bet,
		CreatedAt: protoEvent.CreatedAt.AsTime(),
		Player1:   p1Data,
		Player2:   p2Data,
	}

	return domainEventData, nil
}
