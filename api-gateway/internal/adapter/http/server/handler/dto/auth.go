package dto

import (
	"api-gateway/internal/model"
	"errors"
	"time"
)

type UserRegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisteredResponse struct {
	ID int64 `json:"id"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

type UpdateUsernameResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeleteByIDResponse struct {
	ID        int64     `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateEmailRequest struct {
	Email string `json:"email"`
}

type ConfirmEmailChangeRequest struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func ToUserFromRegisterRequest(req UserRegisterRequest) (model.User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return model.User{}, errors.New("missing required fields")
	}
	err := validateUserCreateRequest(req)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		Username:        req.Username,
		Email:           req.Email,
		CurrentPassword: req.Password,
	}, nil
}

func FromUserToRegisteredResponse(id int64) UserRegisteredResponse {
	return UserRegisteredResponse{
		ID: id,
	}
}

func ToUserFromLoginRequest(req UserLoginRequest) (model.User, error) {
	if req.Email == "" || req.Password == "" {
		return model.User{}, errors.New("missing required fields")
	}
	err := validateUserLoginRequest(req)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		Email:           req.Email,
		CurrentPassword: req.Password,
	}, nil
}

func FromTokenToLoginResponse(token model.Token) UserLoginResponse {
	return UserLoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}

func ToTokenFromRefreshTokenRequest(req UserRefreshTokenRequest) (model.Token, error) {
	if req.RefreshToken == "" {
		return model.Token{}, errors.New("refresh token is required")
	}
	return model.Token{
		RefreshToken: req.RefreshToken,
	}, nil
}

func FromTokenToRefreshTokenResponse(token model.Token) UserLoginResponse {
	return UserLoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}

func ToUserFromUpdateUsernameRequest(req UpdateUsernameRequest) (model.UserUpdateData, error) {
	if req.Username == "" {
		return model.UserUpdateData{}, errors.New("username is required")
	}
	return model.UserUpdateData{
		Username: &req.Username,
	}, nil
}

func FromUserToUpdateUsernameResponse(data model.UserUpdateData) UpdateUsernameResponse {
	return UpdateUsernameResponse{
		ID:        *data.ID,
		Username:  *data.Username,
		UpdatedAt: *data.UpdatedAt,
	}
}

func FromUserToDeleteByIDResponse(data model.UserUpdateData) DeleteByIDResponse {
	return DeleteByIDResponse{
		ID:        *data.ID,
		UpdatedAt: *data.UpdatedAt,
	}
}

func ToUserFromUpdateEmailRequest(req UpdateEmailRequest) (model.UserUpdateData, error) {
	if req.Email == "" {
		return model.UserUpdateData{}, errors.New("email is required")
	}
	return model.UserUpdateData{
		Email: &req.Email,
	}, nil
}

func ToRequestToChangeFromConfirmEmailChangeRequest(req ConfirmEmailChangeRequest) (model.RequestToChange, error) {
	if req.Token == "" {
		return model.RequestToChange{}, errors.New("token is required")
	}
	return model.RequestToChange{
		Token: req.Token,
	}, nil
}

func ToUserFromChangePasswordRequest(req ChangePasswordRequest) (model.User, error) {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return model.User{}, errors.New("missing required fields")
	}
	return model.User{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}, nil
}

func ToUserFromRequestPasswordResetRequest(req RequestPasswordResetRequest) (model.User, error) {
	if req.Email == "" {
		return model.User{}, errors.New("email is required")
	}
	return model.User{
		Email: req.Email,
	}, nil
}

func ToRequestToChangeFromResetPasswordRequest(req ResetPasswordRequest) (model.RequestToChange, error) {
	if req.Token == "" || req.NewPassword == "" {
		return model.RequestToChange{}, errors.New("missing required fields")
	}
	return model.RequestToChange{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}, nil
}
