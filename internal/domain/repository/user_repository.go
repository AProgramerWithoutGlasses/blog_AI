package repository

import (
	"grpc-ddd-demo/internal/domain/model/entity"
)

// UserRepository 定义了用户数据访问接口
type UserRepository interface {
	FindByID(id int64) (*entity.User, error)
	Save(user *entity.User) error
}
