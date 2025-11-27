package handlers

import (
	"ecommerce-microservices/product-service/shared/responses"
	"ecommerce-microservices/product-service/usecases"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productUsecase usecases.ProductUsecase
}

func NewProductHandler(productUsecase usecases.ProductUsecase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	category := c.Query("category")

	products, err := h.productUsecase.GetProducts(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.Error("Failed to get products", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Products retrieved successfully", products))
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Invalid product ID", err.Error()))
		return
	}

	product, err := h.productUsecase.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.Error("Product not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Product retrieved successfully", product))
}
