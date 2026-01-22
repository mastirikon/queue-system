package handler

// CreateTaskRequest — упрощённый запрос (только данные уведомления)
type CreateTaskRequest struct {
	OwnerApp  string `json:"owner_app"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Subtext   string `json:"subtext"`
	Messages  string `json:"messages"`
	OtherText string `json:"other_text"`
	Cat       string `json:"cat"`
	NewOnly   string `json:"new_only"`
}
