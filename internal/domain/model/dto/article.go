package dto

type ArticleFirst struct {
	Key      string   `json:"key"`      // 用于标识文章的状态(是否被修改)
	Abstract string   `json:"abstract"` // 发布文章时，提取的文章摘要
	Summary  string   `json:"summary"`  // 发布文章时，提取的文章总结
	Tags     []string `json:"tags"`     // 标签
}

type ArticleSecond struct {
	Abstract string `json:"abstract"` // 发布文章时，提取的文章摘要
	Summary  string `json:"summary"`  // 发布文章时，提取的文章总结
}

type ArticlePrompt struct {
	Content string   // 询问AI的内容
	Tags    []string // 询问AI时提供的标签
}
