package usecase

import (
	"grpc-ddd-demo/internal/domain/model/entity"
	"grpc-ddd-demo/internal/domain/repository"
	"grpc-ddd-demo/internal/domain/service"
)

// UserUseCase 定义用户用例接口
type UserUseCase interface {
	GetUser(id int64) (*entity.User, error)
	CreateUser(user *entity.User) error
}

type userUseCase struct {
	repo          repository.UserRepository
	domainService service.UserDomainService
}

// NewUserUseCase 构造函数
func NewUserUseCase(r repository.UserRepository, ds service.UserDomainService) UserUseCase {
	return &userUseCase{
		repo:          r,
		domainService: ds,
	}
}

func (uc *userUseCase) GetUser(id int64) (*entity.User, error) {
	return uc.repo.FindByID(id)
}

func (uc *userUseCase) CreateUser(user *entity.User) error {
	if err := uc.domainService.ValidateUser(user); err != nil {
		return err
	}
	return uc.repo.Save(user)
}
