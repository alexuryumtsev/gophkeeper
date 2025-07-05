package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
	"github.com/uryumtsevaa/gophkeeper/internal/server/common"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
)

// Handlers обработчики HTTP запросов
type Handlers struct {
	service          interfaces.ServiceInterface
	validationSvc    common.ValidationService
	responseHandler  common.ResponseHandler
}

// withUserValidation middleware для валидации пользователя
func (h *Handlers) withUserValidation(handler func(c *gin.Context, uid uuid.UUID)) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := h.validationSvc.ValidateUserID(c)
		if err != nil {
			return
		}
		handler(c, uid)
	}
}

// withUserAndMasterPassword middleware для валидации пользователя и мастер-пароля
func (h *Handlers) withUserAndMasterPassword(handler func(c *gin.Context, uid uuid.UUID, masterPassword string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := h.validationSvc.ValidateUserID(c)
		if err != nil {
			return
		}
		
		masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
		if err != nil {
			return
		}
		
		handler(c, uid, masterPassword)
	}
}

// withUserSecretAndMasterPassword middleware для валидации пользователя, секрета и мастер-пароля
func (h *Handlers) withUserSecretAndMasterPassword(handler func(c *gin.Context, uid, sid uuid.UUID, masterPassword string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := h.validationSvc.ValidateUserID(c)
		if err != nil {
			return
		}
		
		sid, err := h.validationSvc.ValidateSecretID(c)
		if err != nil {
			return
		}
		
		masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
		if err != nil {
			return
		}
		
		handler(c, uid, sid, masterPassword)
	}
}

// withUserAndSecret middleware для валидации пользователя и секрета
func (h *Handlers) withUserAndSecret(handler func(c *gin.Context, uid, sid uuid.UUID)) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := h.validationSvc.ValidateUserID(c)
		if err != nil {
			return
		}
		
		sid, err := h.validationSvc.ValidateSecretID(c)
		if err != nil {
			return
		}
		
		handler(c, uid, sid)
	}
}

// NewHandlers создает новые обработчики
func NewHandlers(service interfaces.ServiceInterface, validationSvc common.ValidationService, responseHandler common.ResponseHandler) *Handlers {
	return &Handlers{
		service:         service,
		validationSvc:   validationSvc,
		responseHandler: responseHandler,
	}
}

// handleRequest helper для обработки запросов с JSON
func (h *Handlers) handleRequest(c *gin.Context, req any, handler func() (any, error), statusCode int) {
	if req != nil {
		if err := h.validationSvc.BindJSON(c, req); err != nil {
			return
		}
	}

	response, err := handler()
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, statusCode, response)
}

// Register регистрация пользователя
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} models.LoginResponse "Успешная регистрация"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 409 {object} map[string]string "Пользователь уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	h.handleRequest(c, &req, func() (any, error) {
		return h.service.RegisterUser(c.Request.Context(), &req)
	}, http.StatusCreated)
}

// Login авторизация пользователя
// @Summary Авторизация пользователя
// @Description Авторизует пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Данные для авторизации"
// @Success 200 {object} models.LoginResponse "Успешная авторизация"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req models.LoginRequest
	h.handleRequest(c, &req, func() (any, error) {
		return h.service.LoginUser(c.Request.Context(), &req)
	}, http.StatusOK)
}

// CreateSecret создание секрета
// @Summary Создание нового секрета
// @Description Создает новый зашифрованный секрет для авторизованного пользователя
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для шифрования"
// @Param request body models.SecretRequest true "Данные секрета"
// @Success 201 {object} models.SecretResponse "Секрет создан"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets [post]
func (h *Handlers) CreateSecret(c *gin.Context) {
	h.withUserAndMasterPassword(h.createSecretHandler)(c)
}

func (h *Handlers) createSecretHandler(c *gin.Context, uid uuid.UUID, masterPassword string) {
	var req models.SecretRequest
	h.handleRequest(c, &req, func() (any, error) {
		return h.service.CreateSecret(c.Request.Context(), uid, &req, masterPassword)
	}, http.StatusCreated)
}

// GetSecrets получение списка секретов
// @Summary Получение списка секретов
// @Description Возвращает список всех секретов авторизованного пользователя
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для расшифровки"
// @Success 200 {object} models.SecretsListResponse "Список секретов"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets [get]
func (h *Handlers) GetSecrets(c *gin.Context) {
	h.withUserAndMasterPassword(h.getSecretsHandler)(c)
}

func (h *Handlers) getSecretsHandler(c *gin.Context, uid uuid.UUID, masterPassword string) {
	h.handleRequest(c, nil, func() (any, error) {
		return h.service.GetSecrets(c.Request.Context(), uid, masterPassword)
	}, http.StatusOK)
}

// GetSecret получение конкретного секрета
// @Summary Получение секрета по ID
// @Description Возвращает конкретный секрет по его ID
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для расшифровки"
// @Param id path string true "ID секрета" format(uuid)
// @Success 200 {object} models.SecretResponse "Данные секрета"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 404 {object} map[string]string "Секрет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets/{id} [get]
func (h *Handlers) GetSecret(c *gin.Context) {
	h.withUserSecretAndMasterPassword(h.getSecretHandler)(c)
}

func (h *Handlers) getSecretHandler(c *gin.Context, uid, sid uuid.UUID, masterPassword string) {
	h.handleRequest(c, nil, func() (any, error) {
		return h.service.GetSecret(c.Request.Context(), sid, uid, masterPassword)
	}, http.StatusOK)
}

// UpdateSecret обновление секрета
// @Summary Обновление секрета
// @Description Обновляет существующий секрет
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для шифрования"
// @Param id path string true "ID секрета" format(uuid)
// @Param request body models.SecretRequest true "Новые данные секрета"
// @Success 200 {object} models.SecretResponse "Секрет обновлен"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 404 {object} map[string]string "Секрет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets/{id} [put]
func (h *Handlers) UpdateSecret(c *gin.Context) {
	h.withUserSecretAndMasterPassword(h.updateSecretHandler)(c)
}

func (h *Handlers) updateSecretHandler(c *gin.Context, uid, sid uuid.UUID, masterPassword string) {
	var req models.SecretRequest
	h.handleRequest(c, &req, func() (any, error) {
		return h.service.UpdateSecret(c.Request.Context(), sid, uid, &req, masterPassword)
	}, http.StatusOK)
}

// DeleteSecret удаление секрета
// @Summary Удаление секрета
// @Description Удаляет секрет по его ID
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param id path string true "ID секрета" format(uuid)
// @Success 204 "Секрет удален"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 404 {object} map[string]string "Секрет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets/{id} [delete]
func (h *Handlers) DeleteSecret(c *gin.Context) {
	h.withUserAndSecret(h.deleteSecretHandler)(c)
}

func (h *Handlers) deleteSecretHandler(c *gin.Context, uid, sid uuid.UUID) {
	h.handleRequest(c, nil, func() (any, error) {
		err := h.service.DeleteSecret(c.Request.Context(), sid, uid)
		return nil, err
	}, http.StatusNoContent)
}

// SyncSecrets синхронизация секретов
// @Summary Синхронизация секретов
// @Description Синхронизирует секреты между клиентом и сервером
// @Tags sync
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для обработки данных"
// @Param request body models.SyncRequest true "Данные для синхронизации"
// @Success 200 {object} models.SyncResponse "Результат синхронизации"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /sync [post]
func (h *Handlers) SyncSecrets(c *gin.Context) {
	h.withUserAndMasterPassword(h.syncSecretsHandler)(c)
}

func (h *Handlers) syncSecretsHandler(c *gin.Context, uid uuid.UUID, masterPassword string) {
	var req models.SyncRequest
	h.handleRequest(c, &req, func() (any, error) {
		return h.service.SyncSecrets(c.Request.Context(), uid, &req, masterPassword)
	}, http.StatusOK)
}