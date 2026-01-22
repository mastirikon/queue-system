package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
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
	validator   *validator.Validate
}

// NewTaskHandler создаёт новый TaskHandler
func NewTaskHandler(queueClient *queue.Client, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		queueClient: queueClient,
		logger:      logger,
		validator:   validator.New(),
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

	// Валидация
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn("Request validation failed",
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	// Создаём задачу
	task := &domain.Task{
		ID:        uuid.New().String(),
		URL:       req.URL,
		Method:    req.Method,
		Headers:   req.Headers,
		Body:      req.Body,
		CreatedAt: time.Now(),
	}

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
