package usecase

import (
	"auth_svc/internal/model"
	"auth_svc/pkg/def"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

func (uc *User) ChangePassword(ctx context.Context, user model.User) error {
	currentUser, err := uc.repo.GetWithFilter(ctx, model.UserFilter{ID: &user.ID})
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := uc.passwordManager.CheckPassword(currentUser.PasswordHash, user.CurrentPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if user.CurrentPassword == user.NewPassword {
		return fmt.Errorf("new password must differ from current password")
	}

	newHash, err := uc.passwordManager.HashPassword(user.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	update := model.UserUpdateData{
		ID:           &user.ID,
		PasswordHash: &newHash,
		UpdatedAt:    def.Pointer(time.Now()),
	}

	if err := uc.repo.PatchByID(ctx, &update); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (uc *User) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := uc.repo.GetWithFilter(ctx, model.UserFilter{Email: &email})
	if err != nil {
		return fmt.Errorf("user not found with email: %w", err)
	}

	token := uuid.NewString()
	resetToken := model.RequestChangeToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = uc.redisEmailRepo.Save(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	url := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", token)
	subject := "Reset your password"
	body := fmt.Sprintf("Hi!\n\nTo reset your password, click the following link:\n%s\n\nThis link will expire in 15 minutes.", url)

	err = uc.producer.PushEmailChangeRequest(ctx, model.EmailSendRequest{
		To:      email,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

func (uc *User) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := uc.redisEmailRepo.Get(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token: %w", err)
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	hashedPassword, err := uc.passwordManager.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	update := model.UserUpdateData{
		ID:           &resetToken.UserID,
		PasswordHash: &hashedPassword,
		UpdatedAt:    def.Pointer(time.Now()),
	}

	if err := uc.repo.PatchByID(ctx, &update); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := uc.redisEmailRepo.Delete(ctx, token); err != nil {
		log.Printf("warning: failed to delete reset token: %v", err)
	}

	return nil
}
