package subscriber

import (
	"context"
	"github.com/nats-io/nats.go"
	"log"
	"time"
	"user_svc/internal/adapter/nats/subscriber/dto"
)

const PushTimeout = time.Second * 30

type User struct {
	userUsecase UserUsecase
}

func NewUserSubscriber(
	userUsecase UserUsecase,
) *User {
	return &User{
		userUsecase: userUsecase,
	}
}

func (c *User) UserCreatedEvent(ctx context.Context, msg *nats.Msg) error {
	log.Printf("received: %s", string(msg.Data))

	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()
	pbUser, err := dto.ToUserCreated(msg.Data)
	if err != nil {
		return err
	}
	if err := c.userUsecase.RecordUser(ctx, pbUser); err != nil {
		log.Printf("Failed to record order statistics: %v", err)
		return err
	}
	return nil
}

func (c *User) UserUpdatedEvent(ctx context.Context, msg *nats.Msg) error {
	log.Printf("received: %s", string(msg.Data))

	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	userUpdate, err := dto.ToUserUpdated(msg.Data)
	if err != nil {
		return err
	}
	if err := c.userUsecase.UpdateUsername(ctx, *userUpdate); err != nil {
		log.Printf("Failed to update user: %v", err)
		return err
	}
	return nil
}

func (c *User) UserDeletedEvent(ctx context.Context, msg *nats.Msg) error {
	log.Printf("received: %s", string(msg.Data))
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	userDelete, err := dto.ToUserDeleted(msg.Data)
	if err != nil {
		return err
	}
	if err := c.userUsecase.DeleteByID(ctx, *userDelete); err != nil {
		log.Printf("Failed to delete user: %v", err)
		return err
	}
	return nil
}
