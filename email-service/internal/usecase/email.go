package usecase

import (
	"context"
	"email_svc/internal/model"
)

type EmailDetail struct {
	sender EmailPresenter
}

func NewEmailDetail(sender EmailPresenter) *EmailDetail {
	return &EmailDetail{
		sender: sender,
	}
}

func (c *EmailDetail) Send(ctx context.Context, detail model.EmailSentDetail) error {
	return c.sender.Send(ctx, detail)
}
