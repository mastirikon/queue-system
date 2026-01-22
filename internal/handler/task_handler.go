package handler

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mastirikon/queue-system/internal/domain"
	"github.com/mastirikon/queue-system/internal/queue"
	"go.uber.org/zap"
)

// TaskHandler обрабатывает HTTP запросы для задач
type TaskHandler struct {
	queueClient *queue.Client
	logger      *zap.Logger
	targetURL   string
}

// NewTaskHandler создаёт новый TaskHandler
func NewTaskHandler(queueClient *queue.Client, logger *zap.Logger, targetURL string) *TaskHandler {
	return &TaskHandler{
		queueClient: queueClient,
		logger:      logger,
		targetURL:   targetURL,
	}
}

// CreateTask обрабатывает POST /tasks
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	// Парсим JSON из body
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse request body",
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format",
		})
	}

	// Сериализуем данные в JSON для отправки
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		h.logger.Error("Failed to marshal request body",
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "serialization_error",
			Message: "Failed to serialize request",
		})
	}

	// Создаём задачу с фиксированным URL из конфига
	task := &domain.Task{
		ID:        uuid.New().String(),
		URL:       h.targetURL,
		Method:    "POST",
		Headers:   map[string]string{"Content-Type": "application/json"},
		Body:      string(bodyBytes),
		CreatedAt: time.Now(),
	}

	h.logger.Info("Creating task",
		zap.String("task_id", task.ID),
		zap.String("target_url", task.URL),
	)

	// Отправляем в очередь
	if err := h.queueClient.EnqueueTask(c.Context(), task); err != nil {
		h.logger.Error("Failed to enqueue task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "enqueue_failed",
			Message: "Failed to enqueue task",
		})
	}

	// Успешный ответ
	return c.Status(fiber.StatusCreated).JSON(CreateTaskResponse{
		TaskID:  task.ID,
		Message: "Task created successfully",
	})
}
