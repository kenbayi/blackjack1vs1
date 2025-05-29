package handler

import (
	"context"

	"email_svc/internal/model"
)

type EmailUsecase interface {
	Send(ctx context.Context, detail model.EmailSentDetail) error
}
