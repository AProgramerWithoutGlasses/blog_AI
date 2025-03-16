package impl

import (
	"encoding/json"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/redis_utils"
	"siwuai/internal/infrastructure/utils"
	"siwuai/pkg/globals"
	"time"
)

type codeDomainService struct {
	repo        persistence.CodeRepository
	redisClient *redis_utils.RedisClient
	bf          *bloom.BloomFilter
}

func NewCodeDomainService(repo persistence.CodeRepository, redisClient *redis_utils.RedisClient, bf *bloom.BloomFilter) service.CodeDomainService {
	return &codeDomainService{
		repo:        repo,
		redisClient: redisClient,
		bf:          bf,
	}
}

func (s *codeDomainService) ExplainCode(req *dto.CodeReq) (*dto.Code, error) {
	key, err := utils.Hash(req.Question)
	if err != nil {
		fmt.Println("app.ExplainCode() unique.Hash() err:", err)
		return nil, err
	}

	var code *dto.Code

	// 1. 检查布隆过滤器
	if !s.bf.Test([]byte(key)) {
		// 若未命中布隆过滤器
		code, err = s.FetchAndSave(req, key)
		if err != nil {
			return nil, err
		}
		// 记录 history
		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			fmt.Println("保存 history 失败:", err)
		}
		return code, nil
	}

	// 2. 检查 Redis 缓存
	if data, err := s.redisClient.Get(key); err == nil && data != "" {
		code = &dto.Code{}
		if err := json.Unmarshal([]byte(data), code); err != nil {
			fmt.Println("JSON 反序列化失败:", err)
			return nil, err
		} else {
			// 记录 history
			err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
			if err != nil {
				fmt.Println("保存 history 失败:", err)
				return nil, err
			}
			return code, nil
		}
	}

	// 3. 查询 MySQL
	entityCode, ok, err := s.repo.GetCodeByHash(key)
	if err != nil {
		fmt.Println("app.ExplainCode() repo.GetCodeByHash() err:", err)
		return nil, err
	}
	if ok {
		// 保存到 redis
		code = entityCode.CodeToDto()
		s.SaveToRedis(key, code)
		// 记录 history
		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			fmt.Println("保存 history 失败:", err)
		}
		return code, nil
	}

	// 4. MySQL 中没有，调用 API 并保存
	code, err = s.FetchAndSave(req, key)
	if err != nil {
		return nil, err
	}
	// 记录 history
	err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
	if err != nil {
		fmt.Println("保存 history 失败:", err)
	}
	return code, nil
}

// FetchAndSave 从LLM获取数据并保存到 mysql、redis、布隆过滤器
func (s *codeDomainService) FetchAndSave(req *dto.CodeReq, key string) (*dto.Code, error) {
	// 从 llm 获取数据
	explain, err := utils.Generate(globals.CodeAICode, req.Question)
	if err != nil {
		fmt.Println("app.ExplainCode() llm.Generate() err:", err)
		return nil, err
	}

	// 保存到 MySQL 的 code 表
	code := &entity.Code{
		Key:         key,
		Explanation: explain["text"].(string),
		Question:    req.Question,
	}
	code.ID, err = s.repo.SaveCode(code)
	if err != nil {
		fmt.Println("app ExplainCode() repo.SaveCode() err:", err)
		return nil, err
	}

	// 保存到 Redis
	dtoCode := code.CodeToDto()
	s.SaveToRedis(key, dtoCode)

	// 添加到布隆过滤器
	s.bf.Add([]byte(key))

	return dtoCode, nil
}

func (s *codeDomainService) SaveToRedis(key string, code *dto.Code) {
	data, err := json.Marshal(code)
	if err != nil {
		fmt.Println("JSON 序列化失败:", err)
		return
	}
	if err := s.redisClient.Set(key, string(data), 24*time.Hour); err != nil {
		fmt.Println("保存到 Redis 失败:", err)
	}
}
