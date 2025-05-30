package handler

import (
	"net/http"

	"api-gateway/internal/adapter/http/server/handler/dto"
	"github.com/gin-gonic/gin"
)

type Statistics struct {
	uc StatisticsUsecase
}

func NewStatistics(uc StatisticsUsecase) *Statistics {
	return &Statistics{uc: uc}
}

func (h *Statistics) GetGeneralGameStats(ctx *gin.Context) {
	stats, err := h.uc.GetGeneralGameStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.FromModelToGeneralGameStatsResponse(*stats))
}

func (h *Statistics) GetUserGameStats(ctx *gin.Context) {
	userID, err := dto.ToUserGameStatsRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.uc.GetUserGameStats(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.FromModelToUserGameStatsResponse(*stats))
}

func (h *Statistics) GetLeaderboard(ctx *gin.Context) {
	req, err := dto.ToLeaderboardRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderboard, err := h.uc.GetLeaderboard(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.FromModelToLeaderboardResponse(*leaderboard))
}
