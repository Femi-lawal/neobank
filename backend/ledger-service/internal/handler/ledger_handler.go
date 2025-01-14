package handler

import (
	"net/http"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/service"
	apperrors "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type LedgerHandler struct {
	Service *service.LedgerService
}

func NewLedgerHandler(s *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{Service: s}
}

type CreateAccountRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Currency      string `json:"currency" binding:"required,len=3"`
	Type          string `json:"type" binding:"required"`
}

func (h *LedgerHandler) CreateAccount(c *gin.Context) {
	// Get authenticated user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		apperrors.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondWithError(c, apperrors.ErrValidation.WithDetails(err.Error()))
		return
	}

	acc, err := h.Service.CreateAccount(userID, req.AccountNumber, req.Name, req.Currency, pkgAccountType(req.Type))
	if err != nil {
		apperrors.RespondWithError(c, apperrors.ErrInternal.WithMessage(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, acc)
}

func (h *LedgerHandler) ListAccounts(c *gin.Context) {
	// Get authenticated user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		apperrors.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	// Only return accounts belonging to the authenticated user
	accounts, err := h.Service.ListAccountsByUser(userID)
	if err != nil {
		apperrors.RespondWithError(c, apperrors.ErrInternal.WithMessage(err.Error()))
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func pkgAccountType(t string) model.AccountType {
	return model.AccountType(t)
}

type TransactionRequest struct {
	Description string `json:"description"`
	Postings    []struct {
		AccountID string `json:"account_id" binding:"required"`
		Amount    string `json:"amount" binding:"required"`
		Direction int    `json:"direction" binding:"required"`
	} `json:"postings" binding:"required"`
}

func (h *LedgerHandler) PostTransaction(c *gin.Context) {
	// Get authenticated user ID for audit
	userID := middleware.GetUserID(c)
	if userID == "" {
		apperrors.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondWithError(c, apperrors.ErrValidation.WithDetails(err.Error()))
		return
	}

	// Map to service type
	sPostings := make([]service.PostingRequest, len(req.Postings))
	for i, p := range req.Postings {
		sPostings[i] = service.PostingRequest{
			AccountID: p.AccountID,
			Amount:    p.Amount,
			Direction: p.Direction,
		}
	}

	entry, err := h.Service.PostTransaction(req.Description, sPostings)
	if err != nil {
		// Check for specific error types
		if err.Error() == "transaction is not balanced" {
			apperrors.RespondWithError(c, apperrors.ErrValidation.WithMessage(err.Error()))
			return
		}
		apperrors.RespondWithError(c, apperrors.ErrInternal.WithMessage(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, entry)
}
