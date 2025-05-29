package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/proto"

	"auth_svc/internal/adapter/nats/producer/dto"
	"auth_svc/internal/model"
	"auth_svc/pkg/nats"
)

const PushTimeout = time.Second * 30

type User struct {
	natsClient         *nats.Client
	createdSubject     string
	updatedSubject     string
	deletedSubject     string
	emailChangeSubject string
}

func NewUserProducer(
	natsClient *nats.Client,
	createdSubject string,
	updatedSubject string,
	deletedSubject string,
	emailChangeSubject string,
) *User {
	return &User{
		natsClient:         natsClient,
		createdSubject:     createdSubject,
		updatedSubject:     updatedSubject,
		deletedSubject:     deletedSubject,
		emailChangeSubject: emailChangeSubject,
	}
}

func (c *User) PushCreated(ctx context.Context, user model.User) error {
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	pbUser := dto.FromUserToUserCreated(user)
	data, err := proto.Marshal(pbUser)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	err = c.natsClient.Conn.Publish(c.createdSubject, data)
	if err != nil {
		return fmt.Errorf("publish Created: %w", err)
	}
	log.Println("UserCreated event pushed:", user.ID)
	return nil
}

func (c *User) PushUpdated(ctx context.Context, user *model.UserUpdateData) error {
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	pbUser := dto.FromUserToUserUpdated(user)
	data, err := proto.Marshal(pbUser)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	err = c.natsClient.Conn.Publish(c.updatedSubject, data)
	if err != nil {
		return fmt.Errorf("publish Updated: %w", err)
	}
	log.Println("UserUpdated event pushed:", user.ID)
	return nil
}

func (c *User) PushDeleted(ctx context.Context, user *model.UserUpdateData) error {
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	pbUser := dto.FromUserToUserDeleted(user)
	data, err := proto.Marshal(pbUser)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	err = c.natsClient.Conn.Publish(c.deletedSubject, data)
	if err != nil {
		return fmt.Errorf("publish Deleted: %w", err)
	}
	log.Println("UserDeleted event pushed:", user.ID)
	return nil
}
