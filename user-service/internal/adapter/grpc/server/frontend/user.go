package frontend

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	usersvc "user_svc/internal/adapter/grpc/server/frontend/proto/user"

	"user_svc/internal/adapter/grpc/server/frontend/dto"
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

func (c *User) GetBalance(ctx context.Context, req *usersvc.UserIDRequest) (*usersvc.GetBalanceResponse, error) {
	balance, err := c.userUsecase.GetBalance(ctx, req.Id)
	if err != nil {
		return nil, dto.FromError(err)
	}
	return &usersvc.GetBalanceResponse{Balance: balance}, nil
}

func (c *User) AddBalance(ctx context.Context, req *usersvc.BalanceUpdateRequest) (*emptypb.Empty, error) {
	if err := c.userUsecase.AddBalance(ctx, req.Id, req.Balance); err != nil {
		return nil, dto.FromError(err)
	}
	return &emptypb.Empty{}, nil
}

func (c *User) SubtractBalance(ctx context.Context, req *usersvc.BalanceUpdateRequest) (*emptypb.Empty, error) {
	if err := c.userUsecase.SubtractBalance(ctx, req.Id, req.Balance); err != nil {
		return nil, dto.FromError(err)
	}
	return &emptypb.Empty{}, nil
}

func (c *User) GetProfile(ctx context.Context, req *usersvc.UserIDRequest) (*usersvc.UserProfileResponse, error) {
	profile, err := c.userUsecase.GetProfile(ctx, req.Id)
	if err != nil {
		return nil, dto.FromError(err)
	}
	return &usersvc.UserProfileResponse{
		User: dto.FromModelToProtoUser(&profile),
	}, nil
}

func (c *User) UpdateProfile(ctx context.Context, req *usersvc.UpdateProfileRequest) (*emptypb.Empty, error) {
	update := dto.ToUserUpdateDataFromUpdateProfileRequest(req)
	log.Printf("hi: %v", update)
	err := c.userUsecase.UpdateProfile(ctx, update)
	if err != nil {
		return nil, dto.FromError(err)
	}
	return &emptypb.Empty{}, nil
}

func (c *User) GetRating(ctx context.Context, req *usersvc.UserIDRequest) (*usersvc.GetRatingResponse, error) {
	rating, err := c.userUsecase.GetRating(ctx, req.Id)
	if err != nil {
		return nil, dto.FromError(err)
	}
	return &usersvc.GetRatingResponse{Rating: rating}, nil
}

func (c *User) UpdateRating(ctx context.Context, req *usersvc.RatingUpdateResponse) (*emptypb.Empty, error) {
	if err := c.userUsecase.UpdateRating(ctx, req.Id, req.Rating); err != nil {
		return nil, dto.FromError(err)
	}
	return &emptypb.Empty{}, nil
}
