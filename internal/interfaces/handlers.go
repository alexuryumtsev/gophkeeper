package interfaces

import "github.com/gin-gonic/gin"

// HandlerInterface интерфейс для обработчиков
type HandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	CreateSecret(c *gin.Context)
	GetSecrets(c *gin.Context)
	GetSecret(c *gin.Context)
	UpdateSecret(c *gin.Context)
	DeleteSecret(c *gin.Context)
	SyncSecrets(c *gin.Context)
}