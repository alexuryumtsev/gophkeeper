package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"

	"github.com/uryumtsevaa/gophkeeper/docs"
	"github.com/uryumtsevaa/gophkeeper/internal/server/handlers"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
	"github.com/uryumtsevaa/gophkeeper/internal/server/middleware"
)

// Router представляет HTTP роутер
type Router struct {
	engine   *gin.Engine
	handlers *handlers.Handlers
	authSvc  interfaces.AuthServiceInterface
}

// NewRouter создает новый роутер
func NewRouter(handlers *handlers.Handlers, authSvc interfaces.AuthServiceInterface) *Router {
	engine := gin.Default()
	engine.Use(middleware.CORSMiddleware())

	return &Router{
		engine:   engine,
		handlers: handlers,
		authSvc:  authSvc,
	}
}

// SetupRoutes настраивает маршруты
func (r *Router) SetupRoutes(port int) {
	api := r.engine.Group("/api/v1")

	// Публичные маршруты
	auth := api.Group("/auth")
	{
		auth.POST("/register", r.handlers.Register)
		auth.POST("/login", r.handlers.Login)
	}

	// Защищенные маршруты
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(r.authSvc))
	{
		secrets := protected.Group("/secrets")
		{
			secrets.POST("", r.handlers.CreateSecret)
			secrets.GET("", r.handlers.GetSecrets)
			secrets.GET("/:id", r.handlers.GetSecret)
			secrets.PUT("/:id", r.handlers.UpdateSecret)
			secrets.DELETE("/:id", r.handlers.DeleteSecret)
		}

		protected.POST("/sync", r.handlers.SyncSecrets)
	}

	// Healthcheck
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger documentation
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", port)
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// GetEngine возвращает gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

