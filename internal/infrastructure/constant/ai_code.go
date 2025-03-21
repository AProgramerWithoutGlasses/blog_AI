package constant

type AICode string

const (
	ArticleAICode AICode = "article"
	CodeAICode    AICode = "code"
)

type JudgingSignInterface interface {
	GetArticleFlag() AICode
	GetCodeFlag() AICode
}

type judgingSign struct {
}

func NewJudgingSign() JudgingSignInterface {
	return &judgingSign{}
}

func (j *judgingSign) GetArticleFlag() AICode {
	return ArticleAICode
}

func (j *judgingSign) GetCodeFlag() AICode {
	return CodeAICode
}
