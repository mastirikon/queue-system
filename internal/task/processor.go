package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/domain"
	"go.uber.org/zap"
)

// Processor обрабатывает задачи из очереди
type Processor struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewProcessor создаёт новый процессор задач
func NewProcessor(logger *zap.Logger, timeout time.Duration) *Processor {
	return &Processor{
		logger: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProcessHTTPRequest обрабатывает HTTP запрос
func (p *Processor) ProcessHTTPRequest(ctx context.Context, t *asynq.Task) error {
	// Десериализуем payload
	var payload domain.TaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		p.logger.Error("Failed to unmarshal task payload",
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info("Processing task",
		zap.String("task_id", payload.ID),
		zap.String("url", payload.URL),
		zap.String("method", payload.Method),
	)

	// Создаём HTTP запрос
	var bodyReader io.Reader
	if payload.Body != "" {
		bodyReader = bytes.NewBufferString(payload.Body)
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, bodyReader)
	if err != nil {
		p.logger.Error("Failed to create HTTP request",
			zap.String("task_id", payload.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки
	for key, value := range payload.Headers {
		req.Header.Set(key, value)
	}

	// Если есть body, добавляем Content-Type по умолчанию
	if payload.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос
	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Warn("HTTP request failed, will retry",
			zap.String("task_id", payload.ID),
			zap.Error(err),
		)
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа (для логирования)
	respBody, _ := io.ReadAll(resp.Body)

	// Проверяем статус код
	if resp.StatusCode == http.StatusOK {
		p.logger.Info("Task completed successfully",
			zap.String("task_id", payload.ID),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return nil // Задача успешно выполнена
	}

	// Если не 200 OK - возвращаем ошибку для retry
	p.logger.Warn("Task failed with non-200 status, will retry",
		zap.String("task_id", payload.ID),
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(respBody)),
	)

	return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
}
