package handler

import (
	"context"
	"github.com/nats-io/nats.go"
	"log"

	"email_svc/internal/adapter/nats/handler/dto"
)

type Email struct {
	usecase EmailUsecase
}

func NewEmail(usecase EmailUsecase) *Email {
	return &Email{usecase: usecase}
}

func (c *Email) Handler(ctx context.Context, msg *nats.Msg) error {
	client, err := dto.ToEmailDetail(msg)
	if err != nil {
		log.Println("failed to convert Client NATS msg", err)

		return err
	}

	err = c.usecase.Send(ctx, client)
	if err != nil {
		log.Println("failed to create many Bonus", err)

		return err
	}

	return nil
}
