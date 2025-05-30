package producer

import (
	"auth_svc/internal/adapter/nats/producer/dto"
	"auth_svc/internal/model"
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
)

func (c *User) PushEmailChangeRequest(ctx context.Context, req model.EmailSendRequest) error {
	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	pbEmail := dto.FromEmail(req)
	data, err := proto.Marshal(pbEmail)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	err = c.natsClient.Conn.Publish(c.emailChangeSubject, data)
	if err != nil {
		return fmt.Errorf("publish EmailChange: %w", err)
	}

	log.Printf("EmailChange event pushed to %s: %s", c.emailChangeSubject, req.To)
	return nil
}
