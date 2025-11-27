package handlers

import (
	"ecommerce-microservices/auth-service/domain/entities"
	"ecommerce-microservices/auth-service/shared/responses"
	"ecommerce-microservices/auth-service/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase usecases.AuthUsecase
}

func NewAuthHandler(authUsecase usecases.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req entities.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Validation failed", err.Error()))
		return
	}

	user, err := h.authUsecase.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Registration failed", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Success("User registered successfully", user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req entities.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Validation failed", err.Error()))
		return
	}

	loginResp, err := h.authUsecase.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.Error("Login failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Login successful", loginResp))
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.authUsecase.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.Error("User not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Profile retrieved successfully", user))
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req entities.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Validation failed", err.Error()))
		return
	}

	user, err := h.authUsecase.UpdateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error("Update failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Profile updated successfully", user))
}
