package common

import "context"

// Request структура для HTTP запроса
type Request struct {
	Method   string
	Endpoint string
	Body     any
	Response any
	Context  context.Context
	Headers  map[string]string
}

// Handler функция для обработки запроса
type Handler func(*Request) error

// Middleware функция middleware
type Middleware func(Handler) Handler