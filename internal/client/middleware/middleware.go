package middleware

import (
	"fmt"

	"github.com/uryumtsevaa/gophkeeper/internal/client/common"
)

// AuthMiddleware middleware для авторизации
func AuthMiddleware(token string) common.Middleware {
	return func(next common.Handler) common.Handler {
		return func(req *common.Request) error {
			if token != "" {
				if req.Headers == nil {
					req.Headers = make(map[string]string)
				}
				req.Headers["Authorization"] = "Bearer " + token
			}
			return next(req)
		}
	}
}

// MasterPasswordMiddleware middleware для мастер-пароля
func MasterPasswordMiddleware(masterPassword string) common.Middleware {
	return func(next common.Handler) common.Handler {
		return func(req *common.Request) error {
			if masterPassword != "" {
				if req.Headers == nil {
					req.Headers = make(map[string]string)
				}
				req.Headers["X-Master-Password"] = masterPassword
			}
			return next(req)
		}
	}
}

// LoggingMiddleware middleware для логирования
func LoggingMiddleware() common.Middleware {
	return func(next common.Handler) common.Handler {
		return func(req *common.Request) error {
			fmt.Printf("Request: %s %s\n", req.Method, req.Endpoint)
			err := next(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			return err
		}
	}
}

// RetryMiddleware middleware для повторных попыток
func RetryMiddleware(maxRetries int) common.Middleware {
	return func(next common.Handler) common.Handler {
		return func(req *common.Request) error {
			var lastErr error
			for i := 0; i <= maxRetries; i++ {
				err := next(req)
				if err == nil {
					return nil
				}
				lastErr = err
				if i < maxRetries {
					fmt.Printf("Retry %d/%d failed: %v\n", i+1, maxRetries, err)
				}
			}
			return lastErr
		}
	}
}
