package domain

import (
	"encoding/json"
	"time"
)

// Headers представляет HTTP заголовки
type Headers map[string]string

// Task представляет задачу для обработки
type Task struct {
	ID        string    `json:"id"`         // Уникальный ID задачи (UUID)
	URL       string    `json:"url"`        // URL для HTTP запроса
	Method    string    `json:"method"`     // HTTP метод (POST, GET и т.д.)
	Headers   Headers   `json:"headers"`    // HTTP заголовки
	Body      string    `json:"body"`       // Тело запроса (если есть)
	CreatedAt time.Time `json:"created_at"` // Время создания задачи
}

// TaskPayload — это payload для Asynq задачи (что отправляем в Redis)
type TaskPayload struct {
	ID      string  `json:"id"`
	URL     string  `json:"url"`
	Method  string  `json:"method"`
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

// ToPayload конвертирует Task в TaskPayload для Asynq
func (t *Task) ToPayload() ([]byte, error) {
	payload := TaskPayload{
		ID:      t.ID,
		URL:     t.URL,
		Method:  t.Method,
		Headers: t.Headers,
		Body:    t.Body,
	}
	return json.Marshal(payload)
}

// TaskFromPayload создаёт Task из payload
func TaskFromPayload(data []byte) (*TaskPayload, error) {
	var payload TaskPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
