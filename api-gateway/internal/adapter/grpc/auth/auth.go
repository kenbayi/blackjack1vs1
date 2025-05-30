package auth

import (
	svc "api-gateway/internal/adapter/frontend/proto/auth"
	"api-gateway/internal/adapter/grpc/auth/dto"
	"api-gateway/internal/model"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Auth struct {
	auth svc.UserServiceClient
}

func NewAuth(auth svc.UserServiceClient) *Auth {
	return &Auth{
		auth: auth,
	}
}

func (c *Auth) Login(ctx context.Context, request model.User) (model.Token, error) {
	resp, err := c.auth.Login(ctx, &svc.LoginRequest{
		Email:    request.Email,
		Password: request.CurrentPassword,
	})

	if err != nil {
		return model.Token{}, err
	}

	client := dto.FromGRPCClientCreateResponse(resp)

	return client, nil
}

func (c *Auth) Register(ctx context.Context, request model.User) (int64, error) {
	resp, err := c.auth.Register(ctx, &svc.RegisterRequest{
		Username: request.Username,
		Email:    request.Email,
		Password: request.CurrentPassword,
	})
	if err != nil {
		return 0, err
	}
	client := dto.FromGRPCClientRegisterResponse(resp)

	return client, nil
}

func (c *Auth) RefreshToken(ctx context.Context, request model.Token) (model.Token, error) {
	resp, err := c.auth.RefreshToken(ctx, &svc.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
	})
	if err != nil {
		return model.Token{}, err
	}
	client := dto.FromGRPCClientRefreshTokenResponse(resp)

	return client, nil
}

func (c *Auth) DeleteByID(ctx context.Context) (model.UserUpdateData, error) {
	resp, err := c.auth.DeleteByID(ctx, &emptypb.Empty{})
	if err != nil {
		return model.UserUpdateData{}, err
	}
	client := dto.FromGRPCClientDeleteResponse(resp)
	return client, nil
}

func (c *Auth) UpdateUsername(ctx context.Context, request model.UserUpdateData) (model.UserUpdateData, error) {
	resp, err := c.auth.UpdateUsername(ctx, &svc.UpdateUsernameRequest{
		Username: *request.Username,
	})
	if err != nil {
		return model.UserUpdateData{}, err
	}
	client := dto.FromGRPCClientUpdateUsernameResponse(resp)
	return client, nil
}

func (c *Auth) UpdateEmailRequest(ctx context.Context, request model.UserUpdateData) error {
	_, err := c.auth.UpdateEmailRequest(ctx, &svc.UpdateEmailReq{
		Email: *request.Email,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Auth) ConfirmEmailChange(ctx context.Context, request model.RequestToChange) error {
	_, err := c.auth.ConfirmEmailChange(ctx, &svc.EmailChangeReq{
		Token: request.Token,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Auth) ChangePassword(ctx context.Context, request model.User) error {
	_, err := c.auth.ChangePassword(ctx, &svc.ChangePasswordRequest{
		Id:              request.ID,
		CurrentPassword: request.CurrentPassword,
		NewPassword:     request.NewPassword,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Auth) RequestPasswordReset(ctx context.Context, request model.User) error {
	_, err := c.auth.RequestPasswordReset(ctx, &svc.PasswordResetReq{
		Email: request.Email,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Auth) ResetPassword(ctx context.Context, request model.RequestToChange) error {
	_, err := c.auth.ResetPassword(ctx, &svc.ResetPasswordRequest{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	})
	if err != nil {
		return err
	}
	return nil
}
