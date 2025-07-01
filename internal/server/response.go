package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ResponseHandler interface {
	HandleError(c *gin.Context, err error)
	HandleSuccess(c *gin.Context, statusCode int, data interface{})
	HandleValidationError(c *gin.Context, field, message string)
}

type responseHandler struct{}

func NewResponseHandler() ResponseHandler {
	return &responseHandler{}
}

func (r *responseHandler) HandleError(c *gin.Context, err error) {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		c.JSON(validationErr.Status, gin.H{"error": validationErr.Message})
		return
	}

	if strings.Contains(err.Error(), "not found") {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	if strings.Contains(err.Error(), "already exists") {
		c.JSON(http.StatusConflict, gin.H{"error": "Resource already exists"})
		return
	}

	if strings.Contains(err.Error(), "invalid password") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if strings.Contains(err.Error(), "failed to decrypt") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

func (r *responseHandler) HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func (r *responseHandler) HandleValidationError(c *gin.Context, field, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
		"field": field,
	})
}