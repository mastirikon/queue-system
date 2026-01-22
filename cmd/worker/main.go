package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/config"
	"github.com/mastirikon/queue-system/internal/domain"
	"github.com/mastirikon/queue-system/internal/task"
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

	log.Info("Starting Worker service",
		zap.String("env", cfg.Env),
		zap.Int("concurrency", cfg.Worker.Concurrency),
		zap.Duration("retry_interval", cfg.Worker.RetryInterval),
	)

	// Создаём Asynq Server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues: map[string]int{
				"default": 10, // Приоритет очереди
			},
			// Retry с постоянным интервалом 10 секунд
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				return cfg.Worker.RetryInterval
			},
			Logger: newZapLogger(log),
		},
	)

	// Создаём процессор задач с задержкой между задачами
	processor := task.NewProcessor(log, cfg.Worker.RequestTimeout, cfg.Worker.DelayBetweenTask)

	// Регистрируем обработчики
	mux := asynq.NewServeMux()
	mux.HandleFunc(domain.TypeHTTPRequest, processor.ProcessHTTPRequest)

	// Запускаем worker в горутине
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal("Failed to start worker", zap.Error(err))
		}
	}()

	log.Info("Worker started successfully")

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker gracefully...")

	// Graceful shutdown
	srv.Shutdown()

	log.Info("Worker stopped")
}

// newZapLogger создаёт адаптер для Asynq logger
func newZapLogger(log *zap.Logger) asynq.Logger {
	return &zapLogger{logger: log}
}

// zapLogger адаптер для интеграции zap с asynq
type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *zapLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}
