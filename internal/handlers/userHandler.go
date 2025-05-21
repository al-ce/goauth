package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"godiscauth/internal/services"
	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) (*UserHandler, error) {
	if userService == nil {
		return nil, apperrors.ErrUserServiceIsNil
	}
	return &UserHandler{UserService: userService}, nil
}

func (uh *UserHandler) RegisterUser(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	clientIP := c.ClientIP()

	// Expect both email and password
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Bad user registration request")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Attempt registration
	if err := uh.UserService.RegisterUser(body.Email, body.Password); err != nil {
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("User registration failed")

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Registration success
	log.Info().
		Str("email", body.Email).
		Str("clientIP", clientIP).
		Msg("User registration success")

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User %s created", body.Email)})
}

func (uh *UserHandler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	clientIP := c.ClientIP()

	// Expect both email and password
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Bad user registration request")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Attempt login
	tokenString, err := uh.UserService.LoginUser(body.Email, body.Password)
	if err != nil {
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Login failed")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(config.JwtCookieName, tokenString, config.TokenExpiration, "", "", true, true)

	log.Info().
		Str("email", body.Email).
		Str("clientIP", clientIP).
		Msg("login success")
	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
	})
}

func (uh *UserHandler) Logout(c *gin.Context) {
	clientIP := c.ClientIP()

	tokenString, err := c.Cookie(config.JwtCookieName)
	if err != nil {
		log.Info().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Cookie not found")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := uh.UserService.Logout(tokenString); err != nil {
		log.Error().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Logout failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info().
		Str("clientIP", clientIP).
		Msg("Logout success")
	c.SetCookie(config.JwtCookieName, "", -1, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (uh *UserHandler) LogoutEverywhere(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in cookie")
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
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in cookie")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	userID := userIDStr.(string)
	userProfile, err := uh.UserService.GetUserProfile(userID)
	if err != nil {
		log.Info().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("failed to get user profile")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Info().
		Str("clientIP", clientIP).
		Msg("user profile request successful")

	c.JSON(http.StatusOK, gin.H{
		"email":     userProfile.Email,
		"lastLogin": userProfile.LastLogin,
	})
}

// User can permanently delete their account, instead of setting DeletedAt
func (uh *UserHandler) PermanentlyDeleteUser(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in cookie")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}
	userID := userIDStr.(string)
	err := uh.UserService.PermanentlyDeleteUser(userID)
	if err != nil {
		log.Info().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("failed to delete user")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Info().
		Str("clientIP", clientIP).
		Msg("successfully deleted user")
	c.String(http.StatusOK, "account deleted")
}

func (uh *UserHandler) UpdateUser(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in cookie")
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
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Bad user registration request")
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
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Msg("attempt to update user with empty value")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no valid fields provided"})
		return
	}

	if err := uh.UserService.UpdateUser(userID, requestData); err != nil {
		log.Error().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("failed to update user")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info().
		Str("email", body.Email).
		Str("clientIP", clientIP).
		Msg("successfully updated user")
	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}
