package constant

// CacheType 缓存类型
type CacheType string

const (
	CodeCache    CacheType = "code"    // 代码缓存
	ArticleCache CacheType = "article" // 文章缓存
)

type JudgingCacheType interface {
	GetArticleFlag() CacheType
	GetCodeFlag() CacheType
}

type judgingCache struct{}

func NewJudgingCache() JudgingCacheType {
	return &judgingCache{}
}
func (j *judgingCache) GetArticleFlag() CacheType {
	return ArticleCache
}
func (j *judgingCache) GetCodeFlag() CacheType {
	return CodeCache
}
