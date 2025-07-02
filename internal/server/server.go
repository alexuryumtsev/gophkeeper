package server

import (
	"context"
	"fmt"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/router"
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
	service  *Service
	handlers *Handlers
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
	repository := NewPostgresRepository(db.Pool, encryptor)
	authService := NewAuthService(config.JWTSecret)
	
	// Создаем новые сервисы
	responseHandler := NewResponseHandler()
	validationSvc := NewValidationService(responseHandler)
	cryptoSvc := NewCryptoService(encryptor)
	txManager := NewTransactionManager(db.Pool)
	syncSvc := NewSyncService(repository, cryptoSvc)
	
	// Создаем доменные сервисы
	authDomainSvc := NewAuthDomainService(repository, authService)
	secretsDomainSvc := NewSecretsDomainService(repository, cryptoSvc, txManager, syncSvc)
	syncDomainSvc := NewSyncDomainService(syncSvc)
	
	// Группируем доменные сервисы
	domains := DomainServices{
		Auth:    authDomainSvc,
		Secrets: secretsDomainSvc,
		Sync:    syncDomainSvc,
	}
	
	service := NewService(domains)
	handlers := NewHandlers(service, validationSvc, responseHandler)

	// Создаем роутер
	authWrapper := NewAuthServiceWrapper(authService)
	appRouter := router.NewRouter(handlers, authWrapper)
	appRouter.SetupRoutes(config.Port)

	server := &Server{
		config:   config,
		router:   appRouter,
		database: db,
		service:  service,
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
