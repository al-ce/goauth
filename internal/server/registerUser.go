package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"gofit/internal/models"
)

func RegisterUser(c *gin.Context) {
	// Get the request body
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create the user
	user := models.User{Email: body.Email, Password: string(hash)}

	// TODO: Save the user to the database

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User %s created", user.Email)})
}
