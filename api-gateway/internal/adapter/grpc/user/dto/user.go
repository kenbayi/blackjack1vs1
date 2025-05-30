package dto

import (
	svc "api-gateway/internal/adapter/frontend/proto/user"
	"api-gateway/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func FromGRPCGetBalanceResponse(user *svc.GetBalanceResponse) *model.UserProfile {
	return &model.UserProfile{
		Balance: &user.Balance,
	}
}

func FromGRPCGetProfileResponse(user *svc.UserProfileResponse) *model.UserProfile {
	return &model.UserProfile{
		ID:        user.User.Id,
		Email:     user.User.Email,
		Username:  user.User.Username,
		Role:      user.User.Role,
		CreatedAt: *ProtoTimestampToTimePtr(user.User.CreatedAt),
		UpdatedAt: *ProtoTimestampToTimePtr(user.User.UpdatedAt),
		IsDeleted: user.User.IsDeleted,
		Nickname:  &user.User.Nickname,
		Bio:       &user.User.Bio,
		Balance:   &user.User.Balance,
		Rating:    &user.User.Rating,
	}
}

func FromGRPCGetRatingResponse(user *svc.GetRatingResponse) *model.UserProfile {
	return &model.UserProfile{
		Rating: &user.Rating,
	}
}
func ProtoTimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
