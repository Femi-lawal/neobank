package handler

import (
	"net/http"

	"github.com/femi-lawal/new_bank/backend/payment-service/internal/service"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	Service *service.PaymentService
}

func NewPaymentHandler(s *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: s}
}

type TransferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	Description   string `json:"description"`
}

func (h *PaymentHandler) MakeTransfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.Service.InitiateTransfer(req.FromAccountID, req.ToAccountID, req.Amount, req.Currency, req.Description)
	if err != nil {
		// Return 400 or 500 depending on error, but send payment object so user knows it failed
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "payment": payment})
		return
	}

	c.JSON(http.StatusCreated, payment)
}
