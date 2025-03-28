package utils

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/chains"
	"go.uber.org/zap"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"siwuai/internal/domain/model/dto"
)

// Generate 函数
func Generate(flag constant.AICode, value interface{}, cfg config.Config) (answer map[string]any, err error) {
	var promptTemplate prompts.ChatPromptTemplate
	var input map[string]any
	// 初始化LLM
	llm, err := openai.New(
		openai.WithToken(cfg.Llm.ApiKey),
		openai.WithModel(cfg.Llm.Model),
		openai.WithBaseURL(cfg.Llm.BaseURL),
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

// GenerateStream 用于调用AI大模型接口，传入你要提问的问题，返回2个正在写入的chan
func GenerateStream(flag constant.AICode, value interface{}, cfg config.Config) (streamChan1, streamChan2 chan string, err error) {
	fmt.Println("开始调用llm生成新答案, 请稍等......")

	streamChan1 = make(chan string)
	streamChan2 = make(chan string)

	// 初始化 LLM
	llm, err := openai.New(
		openai.WithToken(cfg.Llm.ApiKey),
		openai.WithModel(cfg.Llm.Model),
		openai.WithBaseURL(cfg.Llm.BaseURL),
	)
	if err != nil {
		err = fmt.Errorf("openai.New() err: %v", err)
		return
	}

	// 将模板和输入渲染为最终的提示词
	promptValue, err := setPrompt(flag, value)
	if err != nil {
		err = fmt.Errorf("promptTemplate.Format() err: %v", err)
		return
	}

	// 创建通道用于传递流式输出
	temp := ""

	// 启动 goroutine 处理 LLM 流
	go func() {
		defer close(streamChan1) // 流结束后关闭通道
		defer close(streamChan2) // 流结束后关闭通道

		// 定义流式输出的回调函数
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			temp = string(chunk)

			// 检查是否满足跳过条件
			if temp == "" || temp == "\n\n" {
				return nil // 跳过当前循环
			}

			streamChan1 <- temp // 将每个 chunk 发送到通道
			streamChan2 <- temp // 将每个 chunk 发送到通道
			return nil
		}

		// 调用 LLM 的 Generate 方法，支持流式输出
		ctx := context.Background()
		_, err = llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			promptValue,
			llms.WithStreamingFunc(streamingFunc),
			llms.WithTemperature(cfg.Llm.TemperatureCode),
		)
		if err != nil {
			err = fmt.Errorf("llms.GenerateFromSinglePrompt() err: %v", err)
			// 将错误传递给通道（可选，根据需求处理）
			streamChan1 <- fmt.Sprintf("Error: %v", err)
			streamChan2 <- fmt.Sprintf("Error: %v", err)
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
		err = fmt.Errorf("promptTemplate.Format() err: %v", err)
		return
	}

	return
}
