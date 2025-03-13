package grpc

import (
	"context"
	"gorm.io/gorm"
	"siwuai/proto/user"

	"siwuai/internal/app/usecase"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
)

// userGRPCHandler 实现了 pb.UserServiceServer 接口
type userGRPCHandler struct {
	user.UnimplementedUserServiceServer
	uc usecase.UserUseCase
}

// NewUserGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewUserGRPCHandler(db *gorm.DB) user.UserServiceServer {
	repo := persistence.NewMySQLUserRepository(db)
	ds := service.NewUserDomainService()
	uc := usecase.NewUserUseCase(repo, ds)
	return &userGRPCHandler{uc: uc}
}

func (h *userGRPCHandler) GetUser(ctx context.Context, req *user.UserRequest) (*user.UserResponse, error) {
	user1, err := h.uc.GetUser(req.Id)
	if err != nil {
		return nil, err
	}
	return &user.UserResponse{
		Id:    user1.ID,
		Name:  user1.Name,
		Email: user1.Email,
	}, nil
}

func (h *userGRPCHandler) CreateUser(ctx context.Context, req *user.UserResponse) (*user.UserResponse, error) {
	user := &entity.User{
		ID:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	}
	if err := h.uc.CreateUser(user); err != nil {
		return nil, err
	}
	return req, nil
}
