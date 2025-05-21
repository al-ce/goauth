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

// RegisterUser godoc
// @Summary register a new user
// @Schemes
// @Description Add a new user to the database from a valid email and password
// @Accept json
// @Produce json
// @Param request body models.UserCredentialsRequest true "User registration credentials"
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 400 {object} models.ErrorResponse "response with error field"
// @Failure 500 {object} models.ErrorResponse "response with error field"
// @Router /register [post]
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
		err = apperrors.ErrMissingCredentials
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

// LoginUser godoc
// @Summary login a user
// @Schemes
// @Description Login an existing user with valid email and password
// @Accept json
// @Produce json
// @Param request body models.UserCredentialsRequest true "User login credentials"
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 400 {object} models.ErrorResponse "response with error field"
// @Failure 401 {object} models.ErrorResponse "response with error field"
// @Router /login [post]
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
			Msg("Bad user login request")

		err = apperrors.ErrMissingCredentials
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Attempt login
	sessionToken, err := uh.UserService.LoginUser(body.Email, body.Password)
	if err != nil {
		log.Info().
			Str("email", body.Email).
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Login failed")

		var status int
		if err == apperrors.ErrInvalidLogin {
			status = http.StatusUnauthorized
		} else {
			status = http.StatusBadRequest
		}

		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(config.SessionCookieName, sessionToken, config.SessionExpiration, "", "", true, true)

	log.Info().
		Str("email", body.Email).
		Str("clientIP", clientIP).
		Msg("login success")

	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
	})
}

// Logout godoc
// @Summary logout a user
// @Description Logs out a logged in user by deleting the associated session in the database
// @Produce json
// @Success 200 {object} models.MessageResponse "response with success message"
// @Failure 401 {object} models.ErrorResponse "unauthorized - cookie not found"
// @Failure 500 {object} models.ErrorResponse "response with error field"
// @Router /logout [post]
func (uh *UserHandler) Logout(c *gin.Context) {
	clientIP := c.ClientIP()

	sessionToken, err := c.Cookie(config.SessionCookieName)
	if err != nil {
		log.Info().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Cookie not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := uh.UserService.Logout(sessionToken); err != nil {
		log.Error().
			Str("clientIP", clientIP).
			Str("error", err.Error()).
			Msg("Logout failed")

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(config.SessionCookieName, "", -1, "", "", true, true)

	log.Info().
		Str("clientIP", clientIP).
		Msg("Logout success")

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// LogoutEverywhere godoc
// @Summary End all user sessions
// @Schemes
// @Description Logs out a logged in user on all devices by deleting all sessions associated with that user
// @Produce json
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 401 {object} models.ErrorResponse "response with error field"
// @Failure 500 {object} models.ErrorResponse "response with error field"
// @Router /logouteverywhere [post]
func (uh *UserHandler) LogoutEverywhere(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in context")

		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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

	c.SetCookie(config.SessionCookieName, "", -1, "", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out everywhere"})
}

// WhoAmI godoc
// @Summary Get information about the currently logged in user
// @Schemes
// @Description Get a user's client IP, email, last login time, and user ID (can be extended)
// @Produce json
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 400 {object} models.ErrorResponse "response with error field"
// @Failure 401 {object} models.ErrorResponse "response with error field"
// @Router /whoami [get]
func (uh *UserHandler) WhoAmI(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in context")
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
		"clientIP":  clientIP,
		"email":     userProfile.Email,
		"lastLogin": userProfile.LastLogin,
		"userID":    userID,
	})
}

// UpdateUser godoc
// @Summary update user credentials
// @Schemes
// @Description Update a user's email or password in the database
// @Accept json
// @Produce json
// @Param request body models.UserCredentialsRequest true "User registration credentials"
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 400 {object} models.ErrorResponse "response with error field"
// @Failure 401 {object} models.ErrorResponse "response with error field"
// @Failure 500 {object} models.ErrorResponse "response with error field"
// @Router /updateuser [post]
func (uh *UserHandler) UpdateUser(c *gin.Context) {
	clientIP := c.ClientIP()

	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in context")

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
			Msg("Bad user update request")

		err = apperrors.ErrMissingCredentials
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

// PermanentlyDeleteUser godoc
// @Summary Delete a user
// @Schemes
// @Description Delete a user from the database permanently along with all their sessions
// @Produce json
// @Success 200 {object} models.MessageResponse "response with message field"
// @Failure 401 {object} models.ErrorResponse "response with error field"
// @Failure 500 {object} models.ErrorResponse "response with error field"
// @Router /deleteaccount [DELETE]
func (uh *UserHandler) PermanentlyDeleteUser(c *gin.Context) {
	clientIP := c.ClientIP()
	userIDStr, exists := c.Get("userID")
	if !exists {
		log.Info().
			Str("clientIP", clientIP).
			Msg("userID not found in context")

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

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Account no longer exists, so we can clear cookie
	// NOTE: we are assuming the database will delete all associated sessions once the
	// corresponding user row is deleted
	c.SetCookie(config.SessionCookieName, "", -1, "", "", true, true)

	log.Info().
		Str("clientIP", clientIP).
		Msg("successfully deleted user")

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
