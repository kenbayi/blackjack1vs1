package dto

import (
	eventsproto "auth_svc/internal/adapter/grpc/server/frontend/proto/events"
	"auth_svc/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func FromUserToUserCreated(client model.User) *eventsproto.UserCreated {
	return &eventsproto.UserCreated{
		User: &eventsproto.User{
			Id:        client.ID,
			Email:     client.Email,
			Username:  client.Username,
			CreatedAt: timestamppb.New(client.CreatedAt),
			UpdatedAt: timestamppb.New(client.UpdatedAt),
			IsDeleted: client.IsDeleted,
		},
	}
}

func FromUserToUserUpdated(client *model.UserUpdateData) *eventsproto.UserCreated {
	user := &eventsproto.User{
		Id:        *client.ID,
		UpdatedAt: timestamppb.New(*client.UpdatedAt),
	}

	if client.Username != nil {
		user.Username = *client.Username
	}
	if client.Email != nil {
		user.Email = *client.Email
	}

	return &eventsproto.UserCreated{
		User: user,
	}
}

func FromUserToUserDeleted(client *model.UserUpdateData) *eventsproto.UserDeleted {
	return &eventsproto.UserDeleted{
		Id: *client.ID,
	}
}

func FromEmail(client model.EmailSendRequest) *eventsproto.EmailSendRequest {
	return &eventsproto.EmailSendRequest{
		To:      client.To,
		Body:    client.Body,
		Subject: client.Subject,
	}
}
