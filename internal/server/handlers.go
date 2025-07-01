package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// Handlers обработчики HTTP запросов
type Handlers struct {
	service          ServiceInterface
	validationSvc    ValidationService
	responseHandler  ResponseHandler
}

// NewHandlers создает новые обработчики
func NewHandlers(service ServiceInterface, validationSvc ValidationService, responseHandler ResponseHandler) *Handlers {
	return &Handlers{
		service:         service,
		validationSvc:   validationSvc,
		responseHandler: responseHandler,
	}
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
	if err := h.validationSvc.BindJSON(c, &req); err != nil {
		return
	}

	response, err := h.service.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusCreated, response)
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
	if err := h.validationSvc.BindJSON(c, &req); err != nil {
		return
	}

	response, err := h.service.LoginUser(c.Request.Context(), &req)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusOK, response)
}

// CreateSecret создание секрета
// @Summary Создание нового секрета
// @Description Создает новый зашифрованный секрет для авторизованного пользователя
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param X-Master-Password header string true "Мастер-пароль для шифрования"
// @Param request body models.SecretRequest true "Данные секрета"
// @Success 201 {object} models.SecretResponse "Секрет создан"
// @Failure 400 {object} map[string]string "Неверные данные запроса"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /secrets [post]
func (h *Handlers) CreateSecret(c *gin.Context) {
	uid, err := h.validationSvc.ValidateUserID(c)
	if err != nil {
		return
	}

	var req models.SecretRequest
	if err := h.validationSvc.BindJSON(c, &req); err != nil {
		return
	}

	masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
	if err != nil {
		return
	}

	response, err := h.service.CreateSecret(c.Request.Context(), uid, &req, masterPassword)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusCreated, response)
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
	uid, err := h.validationSvc.ValidateUserID(c)
	if err != nil {
		return
	}

	masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
	if err != nil {
		return
	}

	response, err := h.service.GetSecrets(c.Request.Context(), uid, masterPassword)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusOK, response)
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

	response, err := h.service.GetSecret(c.Request.Context(), sid, uid, masterPassword)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusOK, response)
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
	uid, err := h.validationSvc.ValidateUserID(c)
	if err != nil {
		return
	}

	sid, err := h.validationSvc.ValidateSecretID(c)
	if err != nil {
		return
	}

	var req models.SecretRequest
	if err := h.validationSvc.BindJSON(c, &req); err != nil {
		return
	}

	masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
	if err != nil {
		return
	}

	response, err := h.service.UpdateSecret(c.Request.Context(), sid, uid, &req, masterPassword)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusOK, response)
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
	uid, err := h.validationSvc.ValidateUserID(c)
	if err != nil {
		return
	}

	sid, err := h.validationSvc.ValidateSecretID(c)
	if err != nil {
		return
	}

	err = h.service.DeleteSecret(c.Request.Context(), sid, uid)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusNoContent, nil)
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
	uid, err := h.validationSvc.ValidateUserID(c)
	if err != nil {
		return
	}

	var req models.SyncRequest
	if err := h.validationSvc.BindJSON(c, &req); err != nil {
		return
	}

	masterPassword, err := h.validationSvc.ValidateMasterPassword(c)
	if err != nil {
		return
	}

	response, err := h.service.SyncSecrets(c.Request.Context(), uid, &req, masterPassword)
	if err != nil {
		h.responseHandler.HandleError(c, err)
		return
	}

	h.responseHandler.HandleSuccess(c, http.StatusOK, response)
}