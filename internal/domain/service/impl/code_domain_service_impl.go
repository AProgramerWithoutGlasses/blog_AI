package impl

import (
	"fmt"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/utils"
	"strconv"
)

type codeDomainService struct {
	repo persistence.CodeRepository
}

// NewCodeDomainService 创建领域服务实例
func NewCodeDomainService(repo persistence.CodeRepository) service.CodeDomainService {
	return &codeDomainService{repo: repo}
}

// ValidateUser 校验用户信息
func (s *codeDomainService) ValidateUser(code *entity.Code) error {
	return nil
}

// SaveCode 调用llm生成回复并且将回复存入表中
func (s *codeDomainService) SaveCode(req *dto.CodeReq, key string) (*dto.Code, error) {
	explain, err := utils.Generate(req.Question)
	if err != nil {
		fmt.Println("app.ExplainCode() llm.Generate() err:", err)
		return &dto.Code{}, err
	}

	code1 := entity.Code{
		Key:         key,
		Explanation: explain,
		Question:    req.Question,
	}
	code1.ID, err = s.repo.SaveCode(code1)
	if err != nil {
		fmt.Println("app ExplainCode() repo.SaveCode() err:", err)
		return &dto.Code{}, err
	}

	return code1.CodeToDto(), nil

}

func (s *codeDomainService) ExplainCode(req *dto.CodeReq) (code1 *dto.Code, err error) {
	// 获取hash值
	key, err := utils.Hash(req.Question)
	if err != nil {
		fmt.Println("app.ExplainCode() unique.Hash() err:", err)
		return
	}

	code2, ok, err := s.repo.GetCodeByHash(key)
	if err != nil {
		fmt.Println("app.ExplainCode() repo.GetCodeByHash() err:", err)
		return
	}
	code1 = code2.CodeToDto()

	if !ok {
		code1, err = s.SaveCode(req, key)
		if err != nil {
			fmt.Println("app.ExplainCode() codeDomainService.SaveCode() err:", err)
			return
		}
	}

	history := entity.History{
		UserID: req.UserId,
		CodeID: code1.ID,
	}

	err = s.repo.SaveHistory(history)
	if err != nil {
		fmt.Println("app.ExplainCode() repo.SaveHistory() err:", err)
		return
	}

	history1, err := s.repo.GetHistory(strconv.Itoa(int(req.UserId)))
	if err != nil {
		fmt.Println("app.ExplainCode() repo.GetHistory() err:", err)
		return
	}

	fmt.Println("history:---", history1)
	return
}
