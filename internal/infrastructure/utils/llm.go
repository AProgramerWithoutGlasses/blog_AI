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
			prompts.NewHumanMessagePromptTemplate("请根据以下文章内容提取摘要和总结，并根据给定的标签匹配文章的标签。文章内容如下：\n{{.article}}\n\n标签列表：{{.tags}}\n\n摘要：\n{{.abstract}}\n\n总结：\n{{.summary}}\n\n匹配的标签：\n{{.matchedTags}}", []string{"article", "abstract", "summary", "tags", "matchedTags"}),
		})

		// 格式化输入
		input = map[string]any{
			"article":     a.Content,
			"abstract":    "",
			"summary":     "",
			"tags":        strings.Join(a.Tags, "、"), // 将标签列表转换为字符串
			"matchedTags": "",
		}
	} else if flag == globals.CodeAICode {

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

	//// 定义输出格式化提示词
	//outputTemplate := "摘要：{{.abstract}}\n总结：{{.summary}}\n匹配的标签：{{.matchedTags}}"
	//// 格式化输出
	//formattedOutput := fmt.Sprintf(outputTemplate, result)

	//// 定义输出格式化模板
	//outputTemplate := "摘要：{{.abstract}}\n总结：{{.summary}}\n匹配的标签：{{.matchedTags}}"
	//
	//// 使用 text/template 处理模板
	//tmpl, err := template.New("output").Parse(outputTemplate)
	//if err != nil {
	//	fmt.Println("Error parsing template:", err)
	//	return
	//}
	//
	//// 创建一个字符串缓冲区
	//var output strings.Builder
	//
	//// 执行模板并写入缓冲区
	//err = tmpl.Execute(&output, result)
	//if err != nil {
	//	fmt.Println("Error executing template:", err)
	//	return
	//}
	//
	//// 获取格式化后的输出
	//formattedOutput := output.String()

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
