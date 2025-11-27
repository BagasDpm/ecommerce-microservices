package handlers

import (
	"ecommerce-microservices/product-service/domain/entities"
	"ecommerce-microservices/product-service/shared/responses"
	"ecommerce-microservices/product-service/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderUsecase usecases.OrderUsecase
}

func NewOrderHandler(orderUsecase usecases.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req entities.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Validation failed", err.Error()))
		return
	}

	order, err := h.orderUsecase.CreateOrder(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Failed to create order", err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, responses.Success("Order is being processed", order))
}
