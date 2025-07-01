package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ValidationService interface {
	ValidateUserID(c *gin.Context) (uuid.UUID, error)
	ValidateMasterPassword(c *gin.Context) (string, error)
	ValidateSecretID(c *gin.Context) (uuid.UUID, error)
	BindJSON(c *gin.Context, v interface{}) error
}

type validationService struct {
	responseHandler ResponseHandler
}

func NewValidationService(responseHandler ResponseHandler) ValidationService {
	return &validationService{
		responseHandler: responseHandler,
	}
}

func (v *validationService) ValidateUserID(c *gin.Context) (uuid.UUID, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "user_id",
			Message: "Unauthorized",
			Status:  http.StatusUnauthorized,
		})
		return uuid.Nil, &ValidationError{Field: "user_id", Message: "Unauthorized"}
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "user_id",
			Message: "Invalid user ID",
			Status:  http.StatusBadRequest,
		})
		return uuid.Nil, &ValidationError{Field: "user_id", Message: "Invalid user ID"}
	}

	return uid, nil
}

func (v *validationService) ValidateMasterPassword(c *gin.Context) (string, error) {
	masterPassword := c.GetHeader("X-Master-Password")
	if masterPassword == "" {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "master_password",
			Message: "Master password required",
			Status:  http.StatusBadRequest,
		})
		return "", &ValidationError{Field: "master_password", Message: "Master password required"}
	}
	return masterPassword, nil
}

func (v *validationService) ValidateSecretID(c *gin.Context) (uuid.UUID, error) {
	secretIDStr := c.Param("id")
	if secretIDStr == "" {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "secret_id",
			Message: "Secret ID is required",
			Status:  http.StatusBadRequest,
		})
		return uuid.Nil, &ValidationError{Field: "secret_id", Message: "Secret ID is required"}
	}

	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "secret_id",
			Message: "Invalid secret ID",
			Status:  http.StatusBadRequest,
		})
		return uuid.Nil, &ValidationError{Field: "secret_id", Message: "Invalid secret ID"}
	}

	return secretID, nil
}

func (v *validationService) BindJSON(c *gin.Context, v_data interface{}) error {
	if err := c.ShouldBindJSON(v_data); err != nil {
		v.responseHandler.HandleError(c, &ValidationError{
			Field:   "json_body",
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return &ValidationError{Field: "json_body", Message: err.Error()}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
	Status  int
}

func (e *ValidationError) Error() string {
	return e.Message
}