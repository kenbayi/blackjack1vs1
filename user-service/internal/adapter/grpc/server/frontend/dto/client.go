package dto

import (
	"time"
	usersvc "user_svc/internal/adapter/grpc/server/frontend/proto/user"
	"user_svc/internal/model"
	"user_svc/pkg/def"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func FromModelToProtoUser(u *model.User) *usersvc.User {
	user := &usersvc.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
		IsDeleted: u.IsDeleted,
	}

	if u.Nickname != nil {
		user.Nickname = *u.Nickname
	}
	if u.Bio != nil {
		user.Bio = *u.Bio
	}
	if u.Balance != nil {
		user.Balance = *u.Balance
	}
	if u.Rating != nil {
		user.Rating = *u.Rating
	}

	return user
}

func ToUserUpdateDataFromUpdateProfileRequest(req *usersvc.UpdateProfileRequest) model.UserUpdateData {
	update := model.UserUpdateData{
		ID:        &req.Id,
		Bio:       &req.Bio,
		Nickname:  &req.Nickname,
		UpdatedAt: def.Pointer(time.Now()),
	}

	if req.Nickname != "" {
		update.Nickname = &req.Nickname
	}
	if req.Bio != "" {
		update.Bio = &req.Bio
	}

	return update
}
