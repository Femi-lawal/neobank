package handler

import (
	"net/http"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/service"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	Service *service.CardService
}

func NewCardHandler(s *service.CardService) *CardHandler {
	return &CardHandler{Service: s}
}

type IssueCardRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

func (h *CardHandler) IssueCard(c *gin.Context) {
	// Get authenticated user ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		apperrors.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	var req IssueCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondWithError(c, apperrors.ErrValidation.WithDetails(err.Error()))
		return
	}

	card, err := h.Service.IssueCard(userID, req.AccountID)
	if err != nil {
		apperrors.RespondWithError(c, apperrors.ErrInternal.WithMessage(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) ListCards(c *gin.Context) {
	// Get authenticated user ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		apperrors.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	// Only return cards belonging to the authenticated user
	cards, err := h.Service.ListCardsByUser(userID)
	if err != nil {
		apperrors.RespondWithError(c, apperrors.ErrInternal.WithMessage(err.Error()))
		return
	}
	c.JSON(http.StatusOK, cards)
}
