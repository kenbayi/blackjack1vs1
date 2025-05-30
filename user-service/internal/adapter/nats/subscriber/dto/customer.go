package dto

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	eventsproto "user_svc/internal/adapter/grpc/server/frontend/proto/events"
	"user_svc/internal/model"
	"user_svc/pkg/def"
)

func ToUserCreated(data []byte) (model.User, error) {
	var pbUser eventsproto.UserCreated
	if err := proto.Unmarshal(data, &pbUser); err != nil {
		return model.User{}, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return model.User{
		ID:        pbUser.User.Id,
		Email:     pbUser.User.Email,
		Username:  pbUser.User.Username,
		CreatedAt: ToTime(pbUser.User.CreatedAt),
		UpdatedAt: ToTime(pbUser.User.UpdatedAt),
		IsDeleted: pbUser.User.IsDeleted,
	}, nil
}

func ToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}
func ToTimePtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	tm := t.AsTime()
	return &tm
}

func ToUserUpdated(data []byte) (*model.UserUpdateData, error) {
	var pb eventsproto.UserCreated
	if err := proto.Unmarshal(data, &pb); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return &model.UserUpdateData{
		ID:        &pb.User.Id,
		Username:  &pb.User.Username,
		UpdatedAt: ToTimePtr(pb.User.UpdatedAt),
	}, nil
}

func ToUserDeleted(data []byte) (*model.UserUpdateData, error) {
	var pb eventsproto.UserDeleted
	if err := proto.Unmarshal(data, &pb); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return &model.UserUpdateData{
		ID:        &pb.Id,
		IsDeleted: def.Pointer(true),
		UpdatedAt: def.Pointer(time.Now()),
	}, nil
}
