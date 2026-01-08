package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создаёт новый логгер в зависимости от окружения
// env может быть "development" или "production"
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return newProduction()
	}
	return newDevelopment()
}

// NewNop создаёт no-op логгер (для тестов)
func NewNop() *zap.Logger {
	return zap.NewNop()
}

// newProduction создаёт production логгер (JSON, INFO+)
func newProduction() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Настройка формата времени
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Уровень логирования
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	return config.Build(
		zap.AddCaller(),                   // Добавляет информацию о месте вызова
		zap.AddStacktrace(zap.ErrorLevel), // Stacktrace только для ERROR+
	)
}

func newDevelopment() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()

	// Красивый цветной вывод для консоли
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Уровень логирования
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	return config.Build(
		zap.AddCaller(),                   // Добавляет информацию о месте вызова
		zap.AddStacktrace(zap.ErrorLevel), // Stacktrace только для ERROR+
	)
}
