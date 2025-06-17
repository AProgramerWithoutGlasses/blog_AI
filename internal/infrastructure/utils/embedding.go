package utils

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/openai"
	"go.uber.org/zap"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
)

func GenerateVector(flag constant.AICode, value interface{}, cfg config.Config) ([][]float32, error) {
	ctx := context.Background()
	// 初始化LLM
	llm, err := openai.New(
		openai.WithToken(cfg.Embedding.ApiKey),
		openai.WithEmbeddingModel(cfg.Embedding.Model),
		openai.WithBaseURL(cfg.Embedding.BaseURL),
	)
	if err != nil {
		zap.L().Error("openai.New() err:", zap.Error(err))
		return nil, err
	}

	if flag == constant.QuestionVectorCode {
		vector := value.(*dto.VectorPrompt)
		// 生成向量
		embedding, err := llm.CreateEmbedding(ctx, vector.Content)
		if err != nil {
			zap.L().Error("llm.CreateEmbedding() err:", zap.Error(err))
			return nil, err
		}
		return embedding, nil
	}
	return nil, fmt.Errorf("flag %s not support", flag)
}
