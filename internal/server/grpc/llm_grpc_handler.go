package grpc

import (
	"bytes"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	pb "grpc-ddd-demo/proto/llm"
)

// llmGRPCHandler 实现了 pb.LLMServiceServer 接口
type llmGRPCHandler struct {
	pb.UnimplementedLLMServiceServer
	llm llms.LLM
}

// NewLLMGRPCHandler 初始化 LLM 处理器，创建并持有 llms.LLM 实例
func NewLLMGRPCHandler() (pb.LLMServiceServer, error) {
	llmInstance, err := openai.New(
		openai.WithModel("deepseek-r1-250120"),
		openai.WithToken("18e25f60-6aff-418f-96fe-55b8cee6a273"),
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
	)
	if err != nil {
		return nil, err
	}
	return &llmGRPCHandler{
		llm: llmInstance,
	}, nil
}

// Generate 接收 prompt，通过 LLM 生成回复，并返回结果
func (h *llmGRPCHandler) Generate(ctx context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {
	var buffer bytes.Buffer
	_, err := llms.GenerateFromSinglePrompt(
		ctx,
		h.llm,
		req.Prompt,
		llms.WithTemperature(0.8), // todo temperature之后放到config.yaml
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			// 将流式生成的内容累积到 buffer 中
			buffer.Write(chunk)
			return nil
		}),
	)
	if err != nil {
		fmt.Println("Generate() err: ", err)
		return nil, err
	}
	return &pb.GenerateResponse{
		Completion: buffer.String(),
	}, nil
}
