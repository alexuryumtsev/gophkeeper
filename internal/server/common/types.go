package common

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ValidationService interface {
	ValidateUserID(c *gin.Context) (uuid.UUID, error)
	ValidateMasterPassword(c *gin.Context) (string, error)
	ValidateSecretID(c *gin.Context) (uuid.UUID, error)
	BindJSON(c *gin.Context, v interface{}) error
}

type ResponseHandler interface {
	HandleError(c *gin.Context, err error)
	HandleSuccess(c *gin.Context, statusCode int, data interface{})
	HandleValidationError(c *gin.Context, field, message string)
}

type ValidationError struct {
	Field   string
	Message string
	Status  int
}

func (e *ValidationError) Error() string {
	return e.Message
}