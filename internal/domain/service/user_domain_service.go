package service

import (
	"errors"
	"siwuai/internal/domain/model/entity"
)

// UserDomainService 定义领域服务接口
type UserDomainService interface {
	ValidateUser(user *entity.User) error
}

type userDomainService struct{}

// NewUserDomainService 创建领域服务实例
func NewUserDomainService() UserDomainService {
	return &userDomainService{}
}

// ValidateUser 校验用户信息
func (s *userDomainService) ValidateUser(user *entity.User) error {
	if user.Email == "" {
		return errors.New("email 不能为空")
	}
	return nil
}
