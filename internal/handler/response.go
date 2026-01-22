package handler

// ErrorResponse — стандартный ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateTaskResponse — ответ на создание задачи
type CreateTaskResponse struct {
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
}
