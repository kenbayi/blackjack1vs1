package usecase

import (
	"context"

	"email_svc/internal/model"
)

type EmailPresenter interface {
	Send(ctx context.Context, detail model.EmailSentDetail) error
}
