package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gofit/internal/services"
	"gofit/pkg/apperrors"
	"gofit/pkg/config"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (uh *UserHandler) RegisterUser(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uh.UserService.RegisterUser(body.Email, body.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User %s created", body.Email)})
}

func (uh *UserHandler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := uh.UserService.LoginUser(body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidLogin})
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(config.JwtCookieName, tokenString, config.TokenExpiration, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{})
}

func (uh *UserHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	userID := userIDStr.(string)
	user, err := uh.UserService.GetUserProfile(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"email":     user.Email,
			"lastLogin": user.LastLogin,
		})
	}
}
