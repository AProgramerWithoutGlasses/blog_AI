package grpc

import (
	"context"
	"go.uber.org/zap"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/utils"
	pbVector "siwuai/proto/vector"
)

type VectorGrpcHandler struct {
	pbVector.UnimplementedVectorServiceServer
	cfg config.Config
}

func NewVectorGrpcHandler(cfg config.Config) *VectorGrpcHandler {
	return &VectorGrpcHandler{
		cfg: cfg,
	}
}

func (v *VectorGrpcHandler) GetVector(ctx context.Context, req *pbVector.GetVectorRequest) (*pbVector.GetVectorResponse, error) {
	zap.L().Info("GenerateVector called")

	// 构造向量模型请求
	vectorPrompt := &dto.VectorPrompt{
		Content: req.Content,
	}

	// 调用ai生成向量
	vector, err := utils.GenerateVector(constant.QuestionVectorCode, vectorPrompt, v.cfg)
	if err != nil {
		zap.L().Error("AI 生成向量失败", zap.Error(err))
		return nil, err
	}

	// 封装返回值
	res := make([]*pbVector.VectorData, len(vector))
	for i, vec := range vector {
		res[i] = &pbVector.VectorData{Values: vec}
	}

	resp := &pbVector.GetVectorResponse{
		Vector: res,
	}
	return resp, nil
}
