package utils

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/chains"
	"go.uber.org/zap"
	"siwuai/internal/infrastructure/constant"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"siwuai/internal/domain/model/dto"
)

// Generate 用于调用AI大模型接口，传入你要提问的问题，返回AI给的答复。
//func Generate(flag globals.AICode, value interface{}) (totalStr string, err error) {
//	// 初始化 LLM
//	llm, err := openai.New(
//		openai.WithModel("deepseek-r1-250120"),
//		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
//		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
//	)
//	if err != nil {
//		fmt.Println("openai.New() err: ", err)
//		return
//	}
//
//	// 将模板和输入渲染为最终的提示词
//	promptValue, err := setPrompt(flag, value)
//	if err != nil {
//		fmt.Println("promptTemplate.Format() err: ", err)
//		return
//	}
//
//	// 调用 LLM 的 Generate 方法，支持流式输出
//	ctx := context.Background()
//	totalStr, err = llms.GenerateFromSinglePrompt(
//		ctx,
//		llm,
//		promptValue,
//	)
//	if err != nil {
//		fmt.Println("llms.GenerateFromSinglePrompt() err: ", err)
//		return
//	}
//
//	return
//}

// Generate 函数
func Generate(flag constant.AICode, value interface{}) (answer map[string]any, err error) {
	var promptTemplate prompts.ChatPromptTemplate
	var input map[string]any
	// 初始化LLM
	llm, err := openai.New(
		openai.WithModel("deepseek-r1-250120"),
		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
	)
	if err != nil {
		zap.L().Error("openai.New() err:", zap.Error(err))
		//fmt.Println("openai.New() err:", err)
		return
	}

	if flag == constant.ArticleAICode {
		a := value.(*dto.ArticlePrompt)
		// 定义提示词模板
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的技术文章分析助手", []string{}),
			//prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}\n\n摘要：\n{{.abstract}}\n\n总结：\n{{.summary}}\n\n匹配的标签：\n{{.matchedTags}}", []string{"article", "abstract", "summary", "tags", "matchedTags"}),
			prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}}", []string{"article", "tags"}),
		})

		// 格式化输入
		input = map[string]any{
			"article": a.Content,
			"tags":    strings.Join(a.Tags, "、"), // 将标签列表转换为字符串
		}
	} else if flag == constant.CodeAICode {

	} else {
		return nil, fmt.Errorf("flag的值超出范围")
	}

	// 创建链
	chain := chains.NewLLMChain(llm, promptTemplate)

	// 运行链
	ctx := context.Background()
	result, err := chain.Call(ctx, input)
	if err != nil {
		zap.L().Error("chain.Call(ctx, input) : ", zap.Error(err))
		//fmt.Println("Error:", err)
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

// GenerateStream 用于调用AI大模型接口，传入你要提问的问题，返回AI给的答复。返回值1为完整答复，返回值2为流式答复。
func GenerateStream(flag constant.AICode, value interface{}) (totalStr string, streamChan chan string, err error) {
	// 初始化 LLM
	llm, err := openai.New(
		openai.WithModel("deepseek-r1-250120"),
		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
	)
	if err != nil {
		fmt.Println("openai.New() err: ", err)
		return
	}

	// 将模板和输入渲染为最终的提示词
	promptValue, err := setPrompt(flag, value)
	if err != nil {
		fmt.Println("promptTemplate.Format() err: ", err)
		return
	}

	// 创建通道用于传递流式输出
	streamChan = make(chan string)

	// 启动 goroutine 处理 LLM 流
	go func() {
		defer close(streamChan) // 流结束后关闭通道

		// 定义流式输出的回调函数
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			streamChan <- string(chunk) // 将每个 chunk 发送到通道
			return nil
		}

		// 调用 LLM 的 Generate 方法，支持流式输出
		ctx := context.Background()
		totalStr, err = llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			promptValue,
			llms.WithStreamingFunc(streamingFunc),
		)
		if err != nil {
			fmt.Println("llms.GenerateFromSinglePrompt() err: ", err)
			// 将错误传递给通道（可选，根据需求处理）
			streamChan <- fmt.Sprintf("Error: %v", err)
		}
	}()

	return
}

// setPrompt 用于设置提示词
func setPrompt(flag constant.AICode, value interface{}) (promptValue string, err error) {
	var promptTemplate prompts.ChatPromptTemplate
	var input map[string]any

	// 根据 flag 设置提示词模板和输入
	if flag == constant.ArticleAICode {
		a := value.(*dto.ArticlePrompt)
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的技术文章分析助手", []string{}),
			prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}}", []string{"article", "tags"}),
		})
		input = map[string]any{
			"article": a.Content,
			"tags":    strings.Join(a.Tags, "、"),
		}
	} else if flag == constant.CodeAICode {
		cp := value.(*dto.CodeReq)
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的代码解释助手", []string{}),
			prompts.NewHumanMessagePromptTemplate("请根据以下{{.language}}代码生成解释，要求解释内容为一段话，字数在300字以内，代码如下：\n{{.code}}", []string{"language", "code"}),
		})
		input = map[string]any{
			"language": cp.CodeType,
			"code":     cp.Question,
		}
	} else {
		fmt.Println("flag的值超出范围")
		return
	}

	// 将模板和输入渲染为最终的提示词
	promptValue, err = promptTemplate.Format(input)
	if err != nil {
		fmt.Println("promptTemplate.Format() err: ", err)
		return
	}

	return
}
