package handler

import (
	"api-gateway/internal/adapter/http/server/handler/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserProfile struct {
	uc UserProfileUsecase
}

func NewUserProfile(uc UserProfileUsecase) *UserProfile {
	return &UserProfile{uc: uc}
}

func (h *UserProfile) GetBalance(ctx *gin.Context) {
	var req dto.GetBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	user, err := dto.ToBalanceFromRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	balance, err := h.uc.GetBalance(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dto.FromModelToGetBalanceResponse(balance))
}

func (h *UserProfile) AddBalance(ctx *gin.Context) {
	var req dto.UpdateBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	user, err := dto.ToUpdateBalanceRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err1 := h.uc.AddBalance(ctx.Request.Context(), user); err1 != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *UserProfile) SubtractBalance(ctx *gin.Context) {
	var req dto.UpdateBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	user, err := dto.ToUpdateBalanceRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err1 := h.uc.SubtractBalance(ctx.Request.Context(), user); err1 != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *UserProfile) GetProfile(ctx *gin.Context) {
	var req dto.GetProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	user, err := dto.ToGetProfileRequest(req)
	log.Printf("GetProfile: going to request: %v", user)
	profile, err := h.uc.GetProfile(ctx.Request.Context(), user)
	log.Printf("GetProfile: from request: %v", profile)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dto.FromModelToUserProfileResponse(profile))
}

func (h *UserProfile) UpdateProfile(ctx *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	profileUpdate := dto.ToUserProfileFromUpdateRequest(req)
	if _, err := h.uc.UpdateProfile(ctx.Request.Context(), profileUpdate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *UserProfile) GetRating(ctx *gin.Context) {
	var req dto.GetRatingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	user, err := dto.ToGetRatingRequest(req)
	rating, err := h.uc.GetRating(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dto.FromModelToGetRatingResponse(rating))
}

func (h *UserProfile) UpdateRating(ctx *gin.Context) {
	var req dto.RatingUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}
	user, err := dto.ToRatingUpdateRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if _, err := h.uc.UpdateRating(ctx.Request.Context(), user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{})
}
