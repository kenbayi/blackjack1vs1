package usecase

import (
	"auth_svc/internal/model"
	"auth_svc/pkg/def"
	"auth_svc/pkg/security"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/mail"
	"time"
)

func (uc *User) UpdateEmailRequest(ctx context.Context, newEmail string) error {
	tokenToVerify, ok := security.TokenFromCtx(ctx)
	if !ok {
		return model.ErrInvalidID
	}
	claims, err := uc.jwtManager.Verify(tokenToVerify)
	if err != nil {
		return model.ErrInvalidID
	}
	rawID, ok := claims["user_id"].(float64)
	if !ok {
		return model.ErrInvalidID
	}
	userID := int64(rawID)

	_, err = mail.ParseAddress(newEmail)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	existing, err := uc.repo.GetWithFilter(ctx, model.UserFilter{Email: &newEmail})
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}
	log.Printf("existing email: %v", existing)
	if existing.Email != "" {
		return fmt.Errorf("email already in use")
	}
	currentData, err := uc.repo.GetWithFilter(ctx, model.UserFilter{ID: &userID})
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}
	if currentData.Email == newEmail {
		return fmt.Errorf("new email is the same as the current one")
	}

	token := uuid.NewString()
	emailToken := model.RequestChangeToken{
		Token:     token,
		UserID:    userID,
		NewEmail:  newEmail,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	err = uc.redisEmailRepo.Save(ctx, emailToken)
	if err != nil {
		return fmt.Errorf("failed to save email confirmation token: %w", err)
	}
	url := fmt.Sprintf("http://localhost:5173/confirm-email?token=%s", token)
	subjectNew := "Confirm your new email address"
	bodyNew := fmt.Sprintf(
		"Hi!\n\nTo confirm your new email address, please click the following link:\n%s\n\nThis link will expire in 10 minutes.",
		url,
	)
	err = uc.producer.PushEmailChangeRequest(ctx, model.EmailSendRequest{
		To:      newEmail,
		Subject: subjectNew,
		Body:    bodyNew,
	})
	if err != nil {
		return fmt.Errorf("failed to publish email send request to new email: %w", err)
	}

	subjectOld := "Your email is being changed"
	bodyOld := fmt.Sprintf(
		"Hi!\n\nWe received a request to change the email address on your account from %s to %s.\n\nIf this wasn't you, please contact support immediately.\n\nIf it was you, no further action is needed.",
		currentData.Email, newEmail,
	)
	log.Printf("Sending email confirmation email to %s", currentData.Email)

	err = uc.producer.PushEmailChangeRequest(ctx, model.EmailSendRequest{
		To:      currentData.Email,
		Subject: subjectOld,
		Body:    bodyOld,
	})
	if err != nil {
		log.Printf("WARNING: failed to send notification to old email: %v", err)
	}

	return nil
}

func (uc *User) ConfirmEmailChange(ctx context.Context, token string) error {
	emailToken, err := uc.redisEmailRepo.Get(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}
	if time.Now().After(emailToken.ExpiresAt) {
		return errors.New("token expired or already used")
	}

	updatedData := model.UserUpdateData{
		ID:        &emailToken.UserID,
		Email:     &emailToken.NewEmail,
		UpdatedAt: def.Pointer(time.Now()),
	}

	err = uc.repo.PatchByID(ctx, &updatedData)

	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	err = uc.redisEmailRepo.Delete(ctx, token)
	if err != nil {
		log.Printf("failed to delete token: %v", err)
	}

	err = uc.producer.PushUpdated(ctx, &updatedData)
	if err != nil {
		return fmt.Errorf("failed to push email confirmation token: %w", err)
	}

	return nil
}
