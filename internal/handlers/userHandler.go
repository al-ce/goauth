package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"goauth/internal/services"
	"goauth/pkg/apperrors"
	"goauth/pkg/config"
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

func (uh *UserHandler) Logout(c *gin.Context) {
	tokenString, err := c.Cookie(config.JwtCookieName)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := uh.UserService.Logout(tokenString); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(config.JwtCookieName, "", -1, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (uh *UserHandler) LogoutEverywhere(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}
	userID := userIDStr.(string)

	log.Info().
		Str("userID", userID).
		Str("clientIP", c.ClientIP()).
		Str("action", "logout_everywhere").
		Msg("User logged out from all devices")

	if err := uh.UserService.LogoutEverywhere(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(config.JwtCookieName, "", -1, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out everywhere"})
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

	// Only accept email or password
	var body struct {
		Email    string `json:"email,omitempty" binding:"omitempty,email"`
		Password string `json:"password,omitempty" binding:"omitempty,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestData := make(map[string]any)
	if body.Email != "" {
		requestData["email"] = body.Email
	}
	if body.Password != "" {
		requestData["password"] = body.Password
	}

	if len(requestData) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no valid fields provided"})
		return
	}

	if err := uh.UserService.UpdateUser(userID, requestData); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}
