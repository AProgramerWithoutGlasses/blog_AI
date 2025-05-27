package grpc

import (
	"context"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/infrastructure/cache"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/utils"
	pbquestion "siwuai/proto/question"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// questionGRPCHandler 实现 pbquestion.QuestionServiceServer 接口
type questionGRPCHandler struct {
	pbquestion.UnimplementedQuestionServiceServer
	db           *gorm.DB
	cfg          config.Config
	cacheManager *cache.CacheManager
	jc           constant.JudgingCacheType
}

// NewQuestionGRPCHandler 构造函数
func NewQuestionGRPCHandler(db *gorm.DB, cfg config.Config, cacheManager *cache.CacheManager, jc constant.JudgingCacheType) pbquestion.QuestionServiceServer {
	return &questionGRPCHandler{
		db:           db,
		cfg:          cfg,
		cacheManager: cacheManager,
		jc:           jc,
	}
}

// GenerateQuestionTitles 实现 gRPC 方法
func (h *questionGRPCHandler) GenerateQuestionTitles(ctx context.Context, req *pbquestion.GenerateQuestionTitlesRequest) (*pbquestion.GenerateQuestionTitlesResponse, error) {
	zap.L().Info("GenerateQuestionTitles called", zap.String("content", req.Content))

	// 构造 AI 请求参数
	questionPrompt := &dto.QuestionPrompt{
		Content: req.Content,
	}

	// 调用 AI 生成标题和标签
	result, err := utils.Generate(constant.QuestionAICode, questionPrompt, h.cfg)
	if err != nil {
		zap.L().Error("AI 生成标题失败", zap.Error(err))
		return &pbquestion.GenerateQuestionTitlesResponse{
			Status: "failed",
		}, err
	}

	// 解析 AI 返回结果
	titles, _ := result["titles"].([]string)
	tags, _ := result["tags"].([]string)
	key, _ := result["key"].(string)
	if key == "" {
		key = "ai-question-key"
	}
	if len(titles) == 0 {
		titles = []string{"AI生成标题"}
	}
	resp := &pbquestion.GenerateQuestionTitlesResponse{
		Key:    key,
		Titles: titles,
		Total:  int32(len(titles)),
		Status: "success",
		Tags:   tags,
	}
	return resp, nil
}
