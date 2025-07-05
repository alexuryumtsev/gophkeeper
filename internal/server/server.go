package server

import (
	"context"
	"fmt"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/server/auth"
	"github.com/uryumtsevaa/gophkeeper/internal/server/handlers"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
	"github.com/uryumtsevaa/gophkeeper/internal/server/repository"
	"github.com/uryumtsevaa/gophkeeper/internal/server/response"
	"github.com/uryumtsevaa/gophkeeper/internal/server/router"
	"github.com/uryumtsevaa/gophkeeper/internal/server/service"
	"github.com/uryumtsevaa/gophkeeper/internal/server/transaction"
	"github.com/uryumtsevaa/gophkeeper/internal/server/validation"
	"github.com/uryumtsevaa/gophkeeper/internal/storage"
)

// Config конфигурация сервера
type Config struct {
	Port      int
	JWTSecret string
	Database  storage.Config
}

// Server HTTP сервер
type Server struct {
	config   Config
	router   *router.Router
	database *storage.Database
	service  interfaces.ServiceInterface
	handlers *handlers.Handlers
}

// NewServer создает новый сервер
func NewServer(config Config) (*Server, error) {
	// Инициализируем базу данных
	db, err := storage.NewDatabase(context.Background(), config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Выполняем миграции
	if err := db.RunMigrations("./migrations", config.Database); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Инициализируем компоненты
	encryptor := crypto.NewAESEncryptor()
	repository := repository.NewPostgresRepository(db.Pool, encryptor)
	authService := auth.NewAuthService(config.JWTSecret)

	// Создаем новые сервисы
	responseHandler := response.NewResponseHandler()
	validationSvc := validation.NewValidationService(responseHandler)
	cryptoSvc := service.NewCryptoService(encryptor)
	txManager := transaction.NewTransactionManager(db.Pool)
	syncSvc := service.NewSyncService(repository, cryptoSvc)

	// Создаем доменные сервисы
	authServiceAdapter := service.NewAuthServiceAdapter(authService)
	authDomainSvc := service.NewAuthDomainService(repository, authServiceAdapter)
	secretsDomainSvc := service.NewSecretsDomainService(repository, cryptoSvc, txManager, syncSvc)
	syncDomainSvc := service.NewSyncDomainService(syncSvc)

	// Группируем доменные сервисы
	domains := service.DomainServices{
		Auth:    authDomainSvc,
		Secrets: secretsDomainSvc,
		Sync:    syncDomainSvc,
	}

	serviceInstance := service.NewService(domains)
	handlers := handlers.NewHandlers(serviceInstance, validationSvc, responseHandler)

	// Создаем роутер
	authWrapper := auth.NewAuthServiceWrapper(authService)
	appRouter := router.NewRouter(handlers, authWrapper)
	appRouter.SetupRoutes(config.Port)

	server := &Server{
		config:   config,
		router:   appRouter,
		database: db,
		service:  serviceInstance,
		handlers: handlers,
	}

	return server, nil
}

// Start запускает сервер
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("Starting server on %s\n", addr)
	return s.router.GetEngine().Run(addr)
}

// Stop останавливает сервер
func (s *Server) Stop() {
	if s.database != nil {
		s.database.Close()
	}
}
