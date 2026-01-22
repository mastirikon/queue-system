package handler

// CreateTaskRequest — запрос на создание задачи
type CreateTaskRequest struct {
	URL     string            `json:"url" validate:"required,url"`
	Method  string            `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
