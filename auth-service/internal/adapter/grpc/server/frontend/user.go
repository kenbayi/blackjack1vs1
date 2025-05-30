package frontend

import (
	usersvc "auth_svc/internal/adapter/grpc/server/frontend/proto/user"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"

	"auth_svc/internal/adapter/grpc/server/frontend/dto"
)

type User struct {
	usersvc.UnimplementedUserServiceServer

	userUsecase UserUsecase
}

func NewUser(uc UserUsecase) *User {
	return &User{
		userUsecase: uc,
	}
}

func (c *User) Register(ctx context.Context, req *usersvc.RegisterRequest) (*usersvc.RegisterResponse, error) {
	user, err := dto.ToUserFromRegisterRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	log.Printf("here before usecase: %v", user)
	id, err := c.userUsecase.Register(ctx, user)
	if err != nil {
		return nil, dto.FromError(err)
	}
	log.Printf("here after usecase: %v", id)

	return &usersvc.RegisterResponse{Id: id}, nil
}

func (c *User) Login(ctx context.Context, req *usersvc.LoginRequest) (*usersvc.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email or password not provided")
	}

	token, err := c.userUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &usersvc.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (c *User) RefreshToken(
	ctx context.Context, req *usersvc.RefreshTokenRequest,
) (*usersvc.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid refresh token")
	}

	token, err := c.userUsecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &usersvc.RefreshTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (c *User) DeleteByID(ctx context.Context, _ *emptypb.Empty) (*usersvc.DeleteByIDResponse, error) {
	updateData, err := c.userUsecase.DeleteByID(ctx)
	if err != nil {
		return nil, dto.FromError(err)
	}

	return &usersvc.DeleteByIDResponse{
		Id:        *updateData.ID,
		UpdatedAt: dto.ToProtoTimestamp(updateData.UpdatedAt),
	}, nil
}

func (c *User) UpdateUsername(ctx context.Context, req *usersvc.UpdateUsernameRequest) (*usersvc.UpdateUsernameResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username cannot be empty")
	}

	updateData, err := c.userUsecase.UpdateUsername(ctx, req.Username)
	if err != nil {
		return nil, dto.FromError(err)
	}

	return &usersvc.UpdateUsernameResponse{
		Id:        *updateData.ID,
		Username:  *updateData.Username,
		UpdatedAt: dto.ToProtoTimestamp(updateData.UpdatedAt),
	}, nil
}

func (c *User) UpdateEmailRequest(ctx context.Context, req *usersvc.UpdateEmailReq) (*emptypb.Empty, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email cannot be empty")
	}
	println("кетти")
	if err := c.userUsecase.UpdateEmailRequest(ctx, req.Email); err != nil {
		return nil, dto.FromError(err)
	}
	println("келди")

	return &emptypb.Empty{}, nil
}

func (c *User) ConfirmEmailChange(ctx context.Context, req *usersvc.EmailChangeReq) (*emptypb.Empty, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token cannot be empty")
	}
	if err := c.userUsecase.ConfirmEmailChange(ctx, req.Token); err != nil {
		return nil, dto.FromError(err)
	}
	return &emptypb.Empty{}, nil
}

func (c *User) ChangePassword(ctx context.Context, req *usersvc.ChangePasswordRequest) (*emptypb.Empty, error) {
	user, err := dto.ToUserFromChangePasswordRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := c.userUsecase.ChangePassword(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (c *User) RequestPasswordReset(ctx context.Context, req *usersvc.PasswordResetReq) (*emptypb.Empty, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email cannot be empty")
	}

	if err := c.userUsecase.RequestPasswordReset(ctx, req.Email); err != nil {
		return nil, dto.FromError(err)
	}

	return &emptypb.Empty{}, nil
}

func (c *User) ResetPassword(ctx context.Context, req *usersvc.ResetPasswordRequest) (*emptypb.Empty, error) {
	if req.Token == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "token and/or new password cannot be empty")
	}
	if err := c.userUsecase.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
