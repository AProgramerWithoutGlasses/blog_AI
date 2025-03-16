package service

import (
	"fmt"
	"regexp"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/code_infrastructure"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/pkg/globals"
	"strings"
)

type articleDomainService struct {
	repo persistence.ArticleRepositoryInterface
}

func NewArticleDomainService(repo persistence.ArticleRepositoryInterface) service.ArticleDomainServiceInterface {
	return &articleDomainService{
		repo: repo,
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
	answer, err := code_infrastructure.Generate(globals.ArticleAICode, ap)
	if err != nil {
		return nil, fmt.Errorf("(a *articleDomainService) VerifyHash -> %v", err)
	}

	// 提取数据
	articleFirst := a.ParseAnswer(answer["text"].(string))
	//articleFirst := a.ParseAnswer(answer)
	articleFirst.Key = key

	fmt.Println()
	fmt.Println("------------------------------------------------")
	fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Abstract)
	fmt.Println("------------------------------------------------")
	fmt.Println()
	fmt.Println("------------------------------------------------")
	fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Summary)
	fmt.Println("------------------------------------------------")
	fmt.Println()
	fmt.Println("------------------------------------------------")
	fmt.Printf("^^^^^^^^^^^^^^^----------> \n %v \n", articleFirst.Tags)
	fmt.Println("------------------------------------------------")
	fmt.Println()

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
	articleInfo, err := a.repo.GetArticleInfo(articleID)
	if err != nil {
		return nil, err
	}
	return articleInfo.ConvertArticleEntityToDtoSecond(), nil

}

// DelArticleInfo 删除文章信息
func (a *articleDomainService) DelArticleInfo(articleID uint) error {
	err := a.repo.DelArticleInfo(articleID)
	return err
}

// ParseAnswer 解析答案
func (a *articleDomainService) ParseAnswer(answer string) *dto.ArticleFirst {
	meta := dto.ArticleFirst{}

	//// 提取摘要（### 摘要 和 --- 之间的内容）
	//if re := regexp.MustCompile(`(?s)### 摘要\n(.*?)\n---`); re.MatchString(answer) {
	//	fmt.Println("------------------>2 我执行了该方法")
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>3 我执行了该方法")
	//		meta.Abstract = strings.TrimSpace(matches[1])
	//		fmt.Println()
	//		fmt.Println("----------------->")
	//		fmt.Println(strings.TrimSpace(matches[1]))
	//		fmt.Println("----------------->")
	//		fmt.Println()
	//		fmt.Println("+++++++++++++++++>")
	//		fmt.Println(meta.Abstract)
	//		fmt.Println("+++++++++++++++++>")
	//	}
	//}
	//
	//// 提取总结（### 文章总结 和 --- 之间的内容）
	//if re := regexp.MustCompile(`(?s)### 总结\n(.*?)\n---`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>a 我执行了该方法")
	//		meta.Summary = strings.TrimSpace(matches[1])
	//		fmt.Println()
	//		fmt.Println("----------------->")
	//		fmt.Println(strings.TrimSpace(matches[1]))
	//		fmt.Println("----------------->")
	//		fmt.Println()
	//		fmt.Println("+++++++++++++++++>")
	//		fmt.Println(meta.Summary)
	//		fmt.Println("+++++++++++++++++>")
	//	}
	//}
	//
	//// 修复标签提取逻辑
	//if re := regexp.MustCompile(`(?s)### 标签\n\s*([^\n]+)`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>b 我执行了该方法")
	//		tagLine := strings.TrimSpace(matches[1])
	//		tagLine = strings.ReplaceAll(tagLine, "**", "") // 移除加粗符号
	//		tags := strings.Split(tagLine, "、")             // 按中文顿号分割
	//		meta.Tags = tags
	//		fmt.Println()
	//		fmt.Println("----------------->")
	//		fmt.Println(tags)
	//		fmt.Println("----------------->")
	//		fmt.Println()
	//		fmt.Println("+++++++++++++++++>")
	//		fmt.Println(meta.Tags)
	//		fmt.Println("+++++++++++++++++>")
	//	}
	//}
	//// 提取摘要（匹配 "摘要: " 到 "总结:" 之间的内容）
	//if re := regexp.MustCompile(`(?s)摘要:\s*(.*?)\s*总结:`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>a 我执行了该方法")
	//		meta.Abstract = strings.TrimSpace(matches[1])
	//	}
	//}
	//
	//// 提取总结（匹配 "总结:" 到 "标签:" 之间的内容）
	//if re := regexp.MustCompile(`(?s)总结:\s*(.*?)\s*标签:`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>b 我执行了该方法")
	//		meta.Summary = strings.TrimSpace(matches[1])
	//	}
	//}
	//
	//// 提取标签（匹配 "标签: " 后的第一行内容）
	//if re := regexp.MustCompile(`标签:\s*([^\n]+)`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		fmt.Println("------------------>c 我执行了该方法")
	//		tagStr := strings.TrimSpace(matches[1])
	//		tags := strings.Split(tagStr, "、")
	//		meta.Tags = tags
	//	}
	//}
	// 修复点1：使用中文冒号匹配

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
		tagStr = strings.ReplaceAll(tagStr, "**", "")     // 移除 **
		tags := strings.Split(tagStr, "、")
		meta.Tags = tags
	}

	//// 提取摘要（匹配 **摘要：** 到 **总结：** 之间的内容）
	//if re := regexp.MustCompile(`(?s)\*\*摘要：\*\*\s*(.*?)\s*\*\*总结：\*\*`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		meta.Abstract = strings.TrimSpace(matches[1])
	//	}
	//}
	//
	//// 提取总结（匹配 **总结：** 到 **匹配的标签：** 之间的内容）
	//if re := regexp.MustCompile(`(?s)\*\*总结：\*\*\s*(.*?)\s*\*\*匹配的标签：\*\*`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		meta.Summary = strings.TrimSpace(matches[1])
	//	}
	//}
	//
	//// 提取标签（匹配 **匹配的标签：** 后的内容）
	//if re := regexp.MustCompile(`\*\*匹配的标签：\*\*\s*([^\n]+)`); re.MatchString(answer) {
	//	matches := re.FindStringSubmatch(answer)
	//	if len(matches) > 1 {
	//		// 处理英文逗号分割和空格
	//		tagStr := strings.ReplaceAll(matches[1], " ", "") // 去除所有空格
	//		tags := strings.Split(tagStr, ",")                // 按英文逗号分割
	//		meta.Tags = tags
	//	}
	//}

	return &meta
}
