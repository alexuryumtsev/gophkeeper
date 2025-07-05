package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/uryumtsevaa/gophkeeper/internal/client/common"
)

// DoRequest выполняет HTTP запрос
func DoRequest(client *http.Client, baseURL string, req *common.Request) error {
	var reqBody io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequestWithContext(req.Context, req.Method, baseURL+req.Endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errorResp)
		if msg, ok := errorResp["error"].(string); ok {
			return fmt.Errorf("server error: %s", msg)
		}
		return fmt.Errorf("server error: status %d", resp.StatusCode)
	}

	if req.Response != nil {
		if err := json.NewDecoder(resp.Body).Decode(req.Response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
