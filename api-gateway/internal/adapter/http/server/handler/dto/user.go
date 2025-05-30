package dto

import (
	"api-gateway/internal/model"
	"time"
)

type GetBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type GetBalanceRequest struct {
	Id int64 `json:"id"`
}

type UpdateBalanceRequest struct {
	Id      int64 `json:"id"`
	Balance int64 `json:"balance"`
}
type GetRatingRequest struct {
	Id int64 `json:"id"`
}
type UserProfileResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Nickname  *string   `json:"nickname,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	Balance   *int64    `json:"balance,omitempty"`
	Rating    *int64    `json:"rating,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

type UpdateProfileRequest struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}
type GetProfileRequest struct {
	Id int64 `json:"id"`
}
type GetRatingResponse struct {
	Rating int64 `json:"rating"`
}

type RatingUpdateRequest struct {
	Id     int64 `json:"id"`
	Rating int64 `json:"rating"`
}

func FromModelToGetBalanceResponse(user model.UserProfile) GetBalanceResponse {
	return GetBalanceResponse{
		Balance: *user.Balance,
	}
}

func FromModelToUserProfileResponse(profile model.UserProfile) UserProfileResponse {
	return UserProfileResponse{
		ID:        profile.ID,
		Username:  profile.Username,
		Email:     profile.Email,
		Role:      profile.Role,
		Nickname:  profile.Nickname,
		Bio:       profile.Bio,
		Balance:   profile.Balance,
		Rating:    profile.Rating,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
		IsDeleted: profile.IsDeleted,
	}
}

func ToUserProfileFromUpdateRequest(req UpdateProfileRequest) model.UserProfile {
	profileUpdate := model.UserProfile{}

	if req.Id != 0 {
		profileUpdate.ID = req.Id
		if req.Nickname != "" {
			profileUpdate.Nickname = &req.Nickname
		}

		if req.Bio != "" {
			profileUpdate.Bio = &req.Bio
		}
	}
	return profileUpdate
}

func FromModelToGetRatingResponse(user model.UserProfile) GetRatingResponse {
	return GetRatingResponse{
		Rating: *user.Rating,
	}
}

func ToBalanceFromRequest(req GetBalanceRequest) (model.UserProfile, error) {
	return model.UserProfile{
		ID: req.Id,
	}, nil
}

func ToUpdateBalanceRequest(req UpdateBalanceRequest) (model.UserProfile, error) {
	return model.UserProfile{
		ID:      req.Id,
		Balance: &req.Balance,
	}, nil
}

func ToGetProfileRequest(req GetProfileRequest) (model.UserProfile, error) {
	return model.UserProfile{
		ID: req.Id,
	}, nil
}

func ToGetRatingRequest(req GetRatingRequest) (model.UserProfile, error) {
	return model.UserProfile{
		ID: req.Id,
	}, nil
}

func ToRatingUpdateRequest(req RatingUpdateRequest) (model.UserProfile, error) {
	return model.UserProfile{
		ID:     req.Id,
		Rating: &req.Rating,
	}, nil
}
