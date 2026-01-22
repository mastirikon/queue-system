package queue

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/domain"
	"go.uber.org/zap"
)

// Client — обёртка над Asynq Client
type Client struct {
	client *asynq.Client
	logger *zap.Logger
}

// NewClient создаёт новый queue client
func NewClient(redisAddr string, logger *zap.Logger) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	return &Client{
		client: client,
		logger: logger,
	}
}

// EnqueueTask отправляет задачу в очередь
func (c *Client) EnqueueTask(ctx context.Context, task *domain.Task) error {
	// Конвертируем Task в payload
	payload, err := task.ToPayload()
	if err != nil {
		c.logger.Error("Failed to marshal task payload",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return err
	}

	// Создаём Asynq задачу
	asynqTask := asynq.NewTask(domain.TypeHTTPRequest, payload)

	// Опции задачи
	opts := []asynq.Option{
		asynq.MaxRetry(8640),            // 24 часа при 10 сек интервале
		asynq.Timeout(30 * time.Second), // Таймаут выполнения задачи
		asynq.Retention(24 * time.Hour), // Хранить 24 часа после завершения
		asynq.TaskID(task.ID),           // Устанавливаем ID задачи
	}

	// Отправляем задачу
	info, err := c.client.EnqueueContext(ctx, asynqTask, opts...)
	if err != nil {
		c.logger.Error("Failed to enqueue task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return err
	}

	c.logger.Info("Task enqueued successfully",
		zap.String("task_id", task.ID),
		zap.String("queue", info.Queue),
		zap.Time("next_process_at", info.NextProcessAt),
	)

	return nil
}

// Close закрывает соединение с Redis
func (c *Client) Close() error {
	return c.client.Close()
}
