package models

type AiRequest struct {
	Topic string `json:"topic" form:"topic" binding:"required"`
}

type AiResponse struct {
	Posts []struct {
		Content string `json:"content"`
	} `json:"posts"`
}
