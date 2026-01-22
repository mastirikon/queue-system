package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mastirikon/queue-system/internal/config"
	"github.com/mastirikon/queue-system/internal/handler"
	"github.com/mastirikon/queue-system/internal/queue"
	pkglogger "github.com/mastirikon/queue-system/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := pkglogger.New(cfg.Env)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting API server",
		zap.String("env", cfg.Env),
		zap.String("host", cfg.API.Host),
		zap.Int("port", cfg.API.Port),
	)

	// Создаём Asynq Client
	queueClient := queue.NewClient(cfg.Redis.Addr, log)
	defer queueClient.Close()

	// Создаём Fiber приложение
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.API.ReadTimeout,
		WriteTimeout: cfg.API.WriteTimeout,
		ErrorHandler: customErrorHandler(log),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Создаём handler с фиксированным URL из конфига
	taskHandler := handler.NewTaskHandler(queueClient, log, cfg.Worker.TargetURL)

	// Роутинг
	api := app.Group("/api/v1")
	api.Post("/tasks", taskHandler.CreateTask)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
		if err := app.Listen(addr); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server gracefully...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server stopped")
}

// customErrorHandler обрабатывает ошибки Fiber
func customErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("Request error",
			zap.Int("status", code),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Error(err),
		)

		return c.Status(code).JSON(fiber.Map{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}
}
