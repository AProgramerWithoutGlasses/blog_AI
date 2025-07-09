package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"strings"
	"time"

	"github.com/tmc/langchaingo/chains"
	"go.uber.org/zap"

	"siwuai/internal/domain/model/dto"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
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
		return
	}

	if flag == constant.ArticleAICode {
		a := value.(*dto.ArticlePrompt)
		// 定义提示词模板
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的技术文章分析助手", []string{}),
			//prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}\n\n摘要：\n{{.abstract}}\n\n总结：\n{{.summary}}\n\n匹配的标签：\n{{.matchedTags}}", []string{"article", "abstract", "summary", "tags", "matchedTags"}),
			prompts.NewHumanMessagePromptTemplate(
				"请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。格式如下"+
					"摘要: \n\n总结: \n\n匹配的标签:"+
					"当文章内容无法识别或为空，未提供有效信息，或提供的文本为无意义字符，无法提取实质性内容或进行总结时，返回nil即可，其他的什么都不需要返回。格式如下"+
					"nil"+
					"文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}}", []string{"article", "tags"}),
		})

		// 格式化输入
		input = map[string]any{
			"article": a.Content,
			"tags":    strings.Join(a.Tags, "、"), // 将标签列表转换为字符串
		}
	} else if flag == constant.CodeAICode {

	} else if flag == constant.QuestionAICode {
		// 新增：处理问题AI生成标题和标签
		q := value.(*dto.QuestionPrompt)
		// 构造prompt和输入参数
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的问题标题和标签生成助手。你必须严格按照指定的JSON格式返回结果，不要添加任何额外的文字说明。", []string{}),
			prompts.NewHumanMessagePromptTemplate(
				"请根据以下问题内容生成3个合适的标题，并为该问题匹配3个相关标签。标签应该是广泛的技术领域或技术框架，例如：'Django'、'React'、'Python'、'Maven'、'Unity'、'Vue.js'、'MySQL'、'Docker'、'Spring Boot'、'机器学习'等，而不是过于细节的具体功能或特性。\n"+
					"你必须严格按照以下JSON格式返回结果，不要添加任何其他内容：\n"+
					"{\n"+
					"  \"titles\": [\"标题1\", \"标题2\", \"标题3\"],\n"+
					"  \"tags\": [\"标签1\", \"标签2\", \"标签3\"]\n"+
					"}\n"+
					"注意：\n"+
					"1. 必须返回3个标题和3个标签\n"+
					"2. 标签必须是广泛的技术领域或框架名称，不要过于细节\n"+
					"3. 不要添加任何额外的说明文字\n"+
					"4. 不要使用反引号包裹JSON\n"+
					"5. 确保返回的是有效的JSON格式\n"+
					"问题内容如下：\n{{.content}}",
				[]string{"content"}),
		})
		input = map[string]any{
			"content": q.Content,
		}
		// 调用LLM
		chain := chains.NewLLMChain(llm, promptTemplate)
		result, err := chain.Call(context.Background(), input)
		if err != nil {
			return nil, err
		}

		// 解析AI返回的JSON字符串
		var aiResponse struct {
			Titles []string `json:"titles"`
			Tags   []string `json:"tags"`
		}

		if resultStr, ok := result["text"].(string); ok {
			// 记录原始返回内容
			zap.L().Info("AI原始返回内容",
				zap.String("raw_response", resultStr))

			// 清理返回的字符串，移除可能的反引号和其他无效字符
			resultStr = strings.TrimSpace(resultStr)
			resultStr = strings.Trim(resultStr, "`")

			// 记录清理后的内容
			zap.L().Info("清理后的内容",
				zap.String("cleaned_response", resultStr))

			// 尝试解析JSON
			if err := json.Unmarshal([]byte(resultStr), &aiResponse); err != nil {
				zap.L().Error("解析AI返回结果失败",
					zap.String("raw_response", resultStr),
					zap.Error(err))
				// 如果解析失败，使用默认值
				aiResponse.Titles = []string{"AI生成标题"}
				aiResponse.Tags = []string{}
			} else {
				// 记录成功解析的内容
				zap.L().Info("成功解析AI返回内容",
					zap.Strings("titles", aiResponse.Titles),
					zap.Strings("tags", aiResponse.Tags))
			}
		} else {
			zap.L().Error("无法获取AI返回的文本内容")
			// 如果无法获取文本结果，使用默认值
			aiResponse.Titles = []string{"AI生成标题"}
			aiResponse.Tags = []string{}
		}

		// 确保至少有一个标题
		if len(aiResponse.Titles) == 0 {
			aiResponse.Titles = []string{"AI生成标题"}
		}

		answer = map[string]any{
			"titles": aiResponse.Titles,
			"tags":   aiResponse.Tags,
			"key":    "ai-question-key",
		}
		return answer, nil
	} else if flag == constant.QuestionAnswerCode {
		// 处理问题AI生成答案
		q := value.(*dto.QuestionPrompt)
		// 构造prompt和输入参数
		promptTemplate = prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewSystemMessagePromptTemplate("你是一个专业的问题回答助手。你必须严格按照指定的JSON格式返回结果，不要添加任何额外的文字说明。", []string{}),
			prompts.NewHumanMessagePromptTemplate(
				"请根据以下问题内容生成一个专业、准确、详细的回答。\n"+
					"你必须严格按照以下JSON格式返回结果，不要添加任何其他内容：\n"+
					"{\n"+
					"  \"answer\": \"你的回答内容\"\n"+
					"}\n"+
					"注意：\n"+
					"1. 回答要专业、准确、详细\n"+
					"2. 不要添加任何额外的说明文字\n"+
					"3. 不要使用反引号包裹JSON\n"+
					"4. 确保返回的是有效的JSON格式\n"+
					"问题内容如下：\n{{.content}}",
				[]string{"content"}),
		})
		input = map[string]any{
			"content": q.Content,
		}
		// 调用LLM
		chain := chains.NewLLMChain(llm, promptTemplate)
		result, err := chain.Call(context.Background(), input)
		if err != nil {
			return nil, err
		}

		// 解析AI返回的JSON字符串
		var aiResponse struct {
			Answer string `json:"answer"`
		}

		if resultStr, ok := result["text"].(string); ok {
			// 记录原始返回内容
			zap.L().Info("AI原始返回内容",
				zap.String("raw_response", resultStr))

			// 清理返回的字符串，移除可能的反引号和其他无效字符
			resultStr = strings.TrimSpace(resultStr)
			resultStr = strings.Trim(resultStr, "`")

			// 记录清理后的内容
			zap.L().Info("清理后的内容",
				zap.String("cleaned_response", resultStr))

			// 尝试解析JSON
			if err := json.Unmarshal([]byte(resultStr), &aiResponse); err != nil {
				zap.L().Error("解析AI返回结果失败",
					zap.String("raw_response", resultStr),
					zap.Error(err))
				return nil, err
			}
		} else {
			zap.L().Error("无法获取AI返回的文本内容")
			return nil, fmt.Errorf("无法获取AI返回的文本内容")
		}

		answer = map[string]any{
			"answer": aiResponse.Answer,
		}
		return answer, nil
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
		return
	}

	var count int
	fmt.Println()
	fmt.Println("------------------>")
	for i, v := range result {
		count++
		fmt.Println(i+"=======", v.(string))
	}
	fmt.Println("------------------>")
	fmt.Println("*********>", count)
	fmt.Println()

	return result, nil
}

// GenerateStream 用于调用AI大模型接口，传入你要提问的问题，返回2个正在写入的chan
func GenerateStream(flag constant.AICode, value interface{}, cfg config.Config) (streamChan1, streamChan2 chan string, err error) {
	fmt.Println("开始调用llm生成新答案, 请稍等......")

	streamChan1 = make(chan string, 1)
	streamChan2 = make(chan string, 1)
	errChan := make(chan error, 1) // 添加错误通道

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

	// 启动 goroutine 处理 LLM 流
	go func() {
		defer close(streamChan1)
		defer close(streamChan2)
		defer close(errChan) // 关闭错误通道

		var temp string
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			temp = string(chunk)
			if temp == "" || temp == "\n\n" {
				return nil
			}
			streamChan1 <- temp
			streamChan2 <- temp
			return nil
		}

		ctx := context.Background()
		_, err = llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			promptValue,
			llms.WithStreamingFunc(streamingFunc),
			llms.WithTemperature(cfg.Llm.TemperatureCode),
		)
		if err != nil {
			// 通过通道将协程中的错误传递给主线程
			errChan <- fmt.Errorf("llms.GenerateFromSinglePrompt() err: %v", err)
			return
		}
		errChan <- nil // 成功时发送 nil
	}()

	// 主线程等待 goroutine 的错误, 为不阻碍后续的运行关联llm生成答案，此处阻塞2s。
	count := 0
	for count < 1 {
		time.Sleep(1 * time.Second)
		select {
		case err = <-errChan:
			return nil, nil, err
		default:
			count++
		}
	}

	return streamChan1, streamChan2, nil
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
			prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。回答中应仅仅只包含三部分: 摘要、总结、匹配的标签，其他多余部分都不要。请尽量选择广泛的技术领域或框架名称作为标签，例如：'Django'、'React'、'Python'、'Maven'、'Unity'等，而不是过于细节的具体功能或特性。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}}", []string{"article", "tags"}),
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
