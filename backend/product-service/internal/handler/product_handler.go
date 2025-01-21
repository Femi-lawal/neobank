package handler

import (
	"net/http"

	"github.com/femi-lawal/new_bank/backend/product-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/product-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Service *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{Service: s}
}

type CreateProductRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Type         string `json:"type" binding:"required"`
	InterestRate string `json:"interest_rate" binding:"required"` // e.g. "0.05"
	Currency     string `json:"currency" binding:"required,len=3"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.Service.CreateProduct(req.Code, req.Name, model.ProductType(req.Type), req.InterestRate, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.Service.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}
