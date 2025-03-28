package grpc

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"siwuai/internal/app"
	impl2 "siwuai/internal/app/impl"
	service "siwuai/internal/domain/service/impl"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/persistence/impl"
	pb "siwuai/proto/article"
)

type articleGRPCHandler struct {
	pb.UnimplementedArticleServiceServer
	repo app.ArticleAppServiceInterface
}

func NewArticleGRPCHandler(db *gorm.DB) pb.ArticleServiceServer {
	repo := impl.NewArticleRepository(db)
	sign := constant.NewJudgingSign()
	ds := service.NewArticleDomainService(repo, sign)
	cr := impl.NewMySQLCodeRepository(db)
	as := impl2.NewArticleAppService(ds, cr)
	return &articleGRPCHandler{
		repo: as,
	}
}

// GetArticleInfoFirst 第一次获取文章的摘要、总结、标签
func (a *articleGRPCHandler) GetArticleInfoFirst(ctx context.Context, req *pb.GetArticleInfoFirstRequest) (*pb.GetArticleInfoFirstResponse, error) {
	articleFirst, err := a.repo.GetArticleInfoFirst(req.Content, req.Tags)
	if err != nil {
		zap.L().Error("GetArticleInfoFirst -> ", zap.Error(err))
		return nil, err
	}
	// 封装数据
	res := &pb.GetArticleInfoFirstResponse{
		Key:      articleFirst.Key,
		Summary:  articleFirst.Summary,
		Abstract: articleFirst.Abstract,
		Tags:     articleFirst.Tags,
	}
	return res, nil
}

// SaveArticleID 保存文章的ID
func (a *articleGRPCHandler) SaveArticleID(ctx context.Context, req *pb.SaveArticleIDRequest) (*pb.SaveArticleIDResponse, error) {
	err := a.repo.SaveArticleID(req.Key, uint(req.ArticleID))
	if err != nil {
		zap.L().Error("SaveArticleID -> ", zap.Error(err))
		return nil, err
	}
	res := &pb.SaveArticleIDResponse{
		Inform: "保存文章ID成功",
	}
	return res, nil
}

// GetArticleInfo 非首次获取文章的信息
func (a *articleGRPCHandler) GetArticleInfo(ctx context.Context, req *pb.GetArticleInfoRequest) (*pb.GetArticleInfoResponse, error) {
	articleSecond, codes, err := a.repo.GetArticleInfo(uint(req.ArticleID), uint(req.UserID))
	if err != nil {
		zap.L().Error("GetArticleInfo -> ", zap.Error(err))
		return nil, err
	}
	res := &pb.GetArticleInfoResponse{
		Summary:  articleSecond.Summary,
		Abstract: articleSecond.Abstract,
	}

	for _, v := range codes {
		value := &pb.Code{
			Question:    v.Question,
			Explanation: v.Explanation,
		}
		res.Codes = append(res.Codes, value)
	}

	return res, nil
}

// DelArticleInfo 删除文章信息
func (a *articleGRPCHandler) DelArticleInfo(ctx context.Context, req *pb.DelArticleInfoRequest) (*pb.DelArticleInfoResponse, error) {
	err := a.repo.DelArticleInfo(uint(req.ArticleID))
	if err != nil {
		zap.L().Error("DelArticleInfo -> ", zap.Error(err))
		return nil, err
	}
	res := &pb.DelArticleInfoResponse{
		Inform: "删除文章信息成功",
	}
	return res, nil
}
