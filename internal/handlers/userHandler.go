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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uh.UserService.RegisterUser(body.Email, body.Password); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := uh.UserService.LoginUser(body.Email, body.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidLogin})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(config.JwtCookieName, tokenString, config.TokenExpiration, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
	})
}

func (uh *UserHandler) GetUserProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	userID := userIDStr.(string)
	userProfile, err := uh.UserService.GetUserProfile(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     userProfile.Email,
		"lastLogin": userProfile.LastLogin,
	})
}

// User can permanently delete their account, instead of setting DeletedAt
func (uh *UserHandler) PermanentlyDeleteUser(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}
	userID := userIDStr.(string)
	err := uh.UserService.PermanentlyDeleteUser(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.String(http.StatusOK, "account deleted")
}

func (uh *UserHandler) UpdateUser(c *gin.Context) {
	// Check auth
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	userID := userIDStr.(string)

	// Check request
	var body struct {
		NewEmail    string `json:"email"`
		NewPassword string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Abort if request contains neither a new password or new email
	if body.NewEmail == "" && body.NewPassword == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "no email or password provided",
		})
	}

	if err := uh.UserService.UpdateUser(userID, body.NewEmail, body.NewPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make(map[string]any)
	if body.NewEmail != "" {
		resp["updatedEmail"] = body.NewEmail
	}
	if body.NewPassword != "" {
		resp["updatedPassword"] = body.NewPassword
	}

	c.JSON(http.StatusOK, resp)
}
