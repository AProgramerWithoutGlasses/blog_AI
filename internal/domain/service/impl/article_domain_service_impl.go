package impl

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"regexp"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/cache"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/utils"
	"strconv"
	"strings"
)

type articleDomainService struct {
	repo persistence.ArticleRepositoryInterface
	sign constant.JudgingSignInterface
	cfg  config.Config
	cm   cache.CacheManagerInterface
	jct  constant.JudgingCacheType
}

func NewArticleDomainService(repo persistence.ArticleRepositoryInterface, sign constant.JudgingSignInterface, cfg config.Config, cm cache.CacheManagerInterface, jct constant.JudgingCacheType) service.ArticleDomainServiceInterface {
	return &articleDomainService{
		repo: repo,
		sign: sign,
		cfg:  cfg,
		cm:   cm,
		jct:  jct,
	}
}

// VerifyHash 验证hash值
func (a *articleDomainService) VerifyHash(key string) (*dto.ArticleFirst, error) {
	articleInfo, err := a.repo.VerifyHash(key)
	if err != nil {
		return nil, err
	}
	return articleInfo.ConvertArticleEntityToDtoFirst(), nil
}

func (a *articleDomainService) AskAI(key string, ap *dto.ArticlePrompt) (*dto.ArticleFirst, error) {
	answer, err := utils.Generate(a.sign.GetArticleFlag(), ap, a.cfg)
	//answer, stream, err := utils.GenerateStream(globals.ArticleAICode, ap)
	if err != nil {
		fmt.Println("utils.Generate() err: ", err)
		return nil, fmt.Errorf("(a *articleDomainService) VerifyHash -> %v", err)
	}

	//fmt.Println("stream:", stream)

	// 提取数据
	articleFirst := a.ParseAnswer(answer["text"].(string))
	//articleFirst := a.ParseAnswer(answer)
	articleFirst.Key = key

	//fmt.Println()
	//fmt.Println("------------------------------------------------")
	//fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Abstract)
	//fmt.Println("------------------------------------------------")
	//fmt.Println()
	//fmt.Println("------------------------------------------------")
	//fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Summary)
	//fmt.Println("------------------------------------------------")
	//fmt.Println()
	//fmt.Println("------------------------------------------------")
	//fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Tags)
	//fmt.Println("------------------------------------------------")
	//fmt.Println()

	// 持久化数据
	articleE := &entity.Article{
		Key:      key,
		Abstract: articleFirst.Abstract,
		Summary:  articleFirst.Summary,
	}

	err = a.repo.SaveArticleInfo(articleE)
	if err != nil {
		return nil, fmt.Errorf("(a *articleDomainService) VerifyHash -> %v", err)
	}

	// 使用 strings.Split 按照 "、" 分割字符串
	//tags := strings.Split(answer["matchedTags"].(string), "、")

	//articleFirst := &dto.ArticleFirst{
	//	Key:      key,
	//	Abstract: answer["abstract"].(string),
	//	Summary:  answer["summary"].(string),
	//	Tags:     tags,
	//}

	return articleFirst, nil
}

// SaveArticleID 保存文章的ID
func (a *articleDomainService) SaveArticleID(key string, articleID uint) error {
	err := a.repo.SaveArticleID(key, articleID)
	return err
}

// GetArticleInfo 非首次获取文章的信息
func (a *articleDomainService) GetArticleInfo(articleID uint) (*dto.ArticleSecond, error) {
	// 从缓存获取数据，优先从本地缓存获取，然后是Redis
	// strconv.FormatUint(uint64(articleID), 10)
	data, err := a.cm.Get(fmt.Sprintf("article:%d", articleID))
	if data == nil && err == nil {
		// 查询数据库
		articleInfo, err := a.repo.GetArticleInfo(articleID)
		if err != nil {
			return nil, err
		}

		articleDto := articleInfo.ConvertArticleEntityToDtoSecond()

		jsonData, err := json.Marshal(*articleDto)
		if err != nil {
			zap.L().Error("文章信息序列化失败，设置Redis缓存失败", zap.Error(err))
		}

		// 设置缓存，同时设置本地缓存和Redis缓存
		a.cm.Set(strconv.FormatUint(uint64(articleInfo.ArticleID), 10), jsonData, a.jct.GetArticleFlag())

		// 返回数据
		return articleDto, nil
	}

	var articleDto dto.ArticleSecond
	err = json.Unmarshal(data, &articleDto)
	if err != nil {
		zap.L().Error("文章信息反序列化失败", zap.Error(err))
		return nil, fmt.Errorf("文章信息反序列化失败 -> %v", err)
	}
	return &articleDto, nil

}

// DelArticleInfo 删除文章信息
func (a *articleDomainService) DelArticleInfo(articleID uint) error {
	err := a.repo.DelArticleInfo(articleID)
	return err
}

// ParseAnswer 解析答案
func (a *articleDomainService) ParseAnswer(answer string) *dto.ArticleFirst {
	meta := dto.ArticleFirst{}

	summaryRe := regexp.MustCompile(`(?s)摘要：\s*(.*?)\s*总结：`)
	if matches := summaryRe.FindStringSubmatch(answer); len(matches) > 1 {
		meta.Abstract = strings.TrimSpace(matches[1])
	}

	// 修复点2：处理中文标点
	conclusionRe := regexp.MustCompile(`(?s)总结：\s*(.*?)\s*匹配的标签：`)
	if matches := conclusionRe.FindStringSubmatch(answer); len(matches) > 1 {
		meta.Summary = strings.TrimSpace(matches[1])
	}

	// 修复点3：支持中文冒号和换行
	tagRe := regexp.MustCompile(`匹配的标签：\s*([^\n]+)`)
	if matches := tagRe.FindStringSubmatch(answer); len(matches) > 1 {
		tagStr := strings.ReplaceAll(matches[1], " ", "") // 移除空格
		tags := strings.Split(tagStr, "、")
		meta.Tags = tags
	}

	return &meta
}
