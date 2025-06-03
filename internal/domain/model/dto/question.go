package dto

// QuestionPrompt 用于AI生成标题和标签的请求参数
type QuestionPrompt struct {
	Content    string `json:"content"`     // 问题正文内容
	QuestionID uint   `json:"question_id"` // 问题ID
}
