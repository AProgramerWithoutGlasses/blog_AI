package utils

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"siwuai/internal/domain/model/dto"
	"siwuai/pkg/globals"
	"strings"
)

//import (
//	"context"
//	"fmt"
//	"github.com/tmc/langchaingo/chains"
//	"github.com/tmc/langchaingo/llms"
//	"github.com/tmc/langchaingo/llms/openai"
//	"github.com/tmc/langchaingo/prompts"
//)
//
//func Generate(prompt string) (answer string, err error) {
//	ctx := context.Background()
//
//	llm, err := openai.New(
//		openai.WithModel("deepseek-r1-250120"),
//		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
//		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
//	)
//	if err != nil {
//		fmt.Println("openai.New() err:", err)
//		return
//	}
//
//	answer, err = llms.GenerateFromSinglePrompt(
//		ctx,
//		llm,
//		prompt,
//		llms.WithTemperature(0.8), // 控制随机性
//		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
//			fmt.Print(string(chunk))
//			return nil
//		}),
//	)
//	if err != nil {
//		fmt.Println("llms.GenerateFromSinglePrompt() err:", err)
//		return
//	}
//
//	return
//}

// Generate 函数
func Generate(flag globals.AICode, value interface{}) (answer map[string]any, err error) {
	var promptTemplate prompts.ChatPromptTemplate
	var input map[string]any
	// 初始化LLM
	llm, err := openai.New(
		openai.WithModel("deepseek-r1-250120"),
		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
	)
	if err != nil {
		fmt.Println("openai.New() err:", err)
		return
	}

	if flag == globals.ArticleAICode {
		a := value.(*dto.ArticlePrompt)
		// 定义提示词模板
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的技术文章分析助手", []string{}),
			//prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}\n\n摘要：\n{{.abstract}}\n\n总结：\n{{.summary}}\n\n匹配的标签：\n{{.matchedTags}}", []string{"article", "abstract", "summary", "tags", "matchedTags"}),
			prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}}", []string{"article", "tags"}),
		})

		//// 格式化输入
		//input = map[string]any{
		//	"article":     a.Content,
		//	"abstract":    "",
		//	"summary":     "",
		//	"tags":        strings.Join(a.Tags, "、"), // 将标签列表转换为字符串
		//	"matchedTags": "",
		//}
		// 格式化输入
		input = map[string]any{
			"article": a.Content,
			"tags":    strings.Join(a.Tags, "、"), // 将标签列表转换为字符串
		}
	} else if flag == globals.CodeAICode {
		// 代码解释功能
		cp := value.(*dto.CodeReq)
		// 定义代码解释的提示词模板
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			// 设置系统消息，确定助手身份
			prompts.NewSystemMessagePromptTemplate("你是一个专业的代码解释助手", []string{}),
			// 设置用户消息，要求生成一段不超过300字的代码解释，且格式为一段话
			prompts.NewHumanMessagePromptTemplate("请根据以下{{.language}}代码生成解释，要求解释内容为一段话，字数在300字以内，代码如下：\n{{.code}}", []string{"language", "code"}),
		})
		input = map[string]any{
			"language": cp.CodeType,
			"code":     cp.Question,
		}
	} else {
		return nil, fmt.Errorf("flag的值超出范围")
	}

	// 创建链
	chain := chains.NewLLMChain(llm, promptTemplate)

	// 运行链
	ctx := context.Background()
	result, err := chain.Call(ctx, input)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var count int
	fmt.Println()
	fmt.Println("------------------>")
	for i, v := range result {
		count++
		fmt.Println(i+"=======", v.(string))
	}
	//fmt.Println(formattedOutput)
	fmt.Println("------------------>")
	fmt.Println("*********>", count)
	fmt.Println()

	return result, nil
}
