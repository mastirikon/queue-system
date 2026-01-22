package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Env string `env:"ENV" envDefault:"development"`

	// API конфигурация
	API APIConfig `envPrefix:"API_"`

	// Worker конфигурация
	Worker WorkerConfig `envPrefix:"WORKER_"`

	// Redis конфигурация
	Redis RedisConfig `envPrefix:"REDIS_"`
}

// APIConfig — настройки API сервиса
type APIConfig struct {
	Port            int           `env:"PORT" envDefault:"8080"`
	Host            string        `env:"HOST" envDefault:"0.0.0.0"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

// WorkerConfig — настройки Worker сервиса
type WorkerConfig struct {
	Concurrency    int           `env:"CONCURRENCY" envDefault:"10"`
	RetryInterval  time.Duration `env:"RETRY_INTERVAL" envDefault:"10s"`
	MaxRetries     int           `env:"MAX_RETRIES" envDefault:"8640"` // 24 часа при 10 сек интервале
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	TargetURL      string        `env:"TARGET_URL" envDefault:"https://tasker-google-sheets.ku-34.netcraze.pro/notify"`
}

// RedisConfig — настройки Redis
type RedisConfig struct {
	Addr     string `env:"ADDR" envDefault:"localhost:6379"`
	Password string `env:"PASSWORD" envDefault:""`
	DB       int    `env:"DB" envDefault:"0"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}
	return config, nil
}
