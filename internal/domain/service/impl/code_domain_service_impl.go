package impl

import (
	"encoding/json"
	"fmt"
	"siwuai/internal/infrastructure/constant"
	"strings"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/redis_utils"
	"siwuai/internal/infrastructure/utils"
)

// 定义锁的过期时间
const lockTTL = 60 * time.Second

type codeDomainService struct {
	repo        persistence.CodeRepository
	redisClient *redis_utils.RedisClient
	bf          *bloom.BloomFilter
	sign        constant.JudgingSignInterface
}

func NewCodeDomainService(repo persistence.CodeRepository, redisClient *redis_utils.RedisClient, bf *bloom.BloomFilter, sign constant.JudgingSignInterface) service.CodeDomainService {
	return &codeDomainService{
		repo:        repo,
		redisClient: redisClient,
		bf:          bf,
		sign:        sign,
	}
}

func (s *codeDomainService) ExplainCode(req *dto.CodeReq) (code *dto.Code, err error) {
	code, err = s.GetAnswer(req)
	if err != nil {
		fmt.Println("s.GetAnswer() ", err)
		return
	}

	if code.ID != 0 {
		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			fmt.Println("s.repo.SaveHistory() ", err)
			return
		}
	}

	return
}

// GetAnswer 用于得到代码解释信息
func (s *codeDomainService) GetAnswer(req *dto.CodeReq) (code *dto.Code, err error) {
	key, err := utils.Hash(req.Question)
	if err != nil {
		fmt.Println("utils.Hash() ", err)
		return nil, err
	}

	lockKey := fmt.Sprintf("lock:%s", key)

	// 尝试获取锁
	locked, err := s.redisClient.TryLock(lockKey, lockTTL)
	if err != nil {
		fmt.Println("TryLock() err: ", err)
		return nil, err
	}

	if locked {
		// 获取到锁，表示我们正在使用该锁
		defer s.redisClient.Unlock(lockKey) // 确保释放锁

		if !s.bf.Test([]byte(key)) {
			fmt.Println("未命中布隆过滤器", key)
			return s.FetchAndSave(req, key)
		}

		fmt.Println("成功命中布隆过滤器，开始查询缓存...")

		if code, err = s.checkRedis(key); err == nil && code != nil {
			return code, nil
		}
		fmt.Printf("未命中redis缓存: %s\n", key)

		if code, err = s.checkMySQL(key); err == nil && code != nil {
			return code, nil
		}
		fmt.Printf("未命中mysql记录: %s\n", key)

		return s.FetchAndSave(req, key)
	} else {
		// 未获取锁，表示该锁正在被别人占用，等待并查询缓存
		for i := 0; i < 60; i++ {
			time.Sleep(1 * time.Second)
			if code, err = s.checkRedis(key); err == nil && code != nil {
				return code, nil
			}
			if code, err = s.checkMySQL(key); err == nil && code != nil {
				return code, nil
			}
		}
		return nil, fmt.Errorf("等待超时，未获取到结果")
	}
}

// checkRedis 检查 Redis 缓存并同步到 MySQL
func (s *codeDomainService) checkRedis(key string) (*dto.Code, error) {
	data, err := s.redisClient.Get(key)
	if err != nil || data == "" {
		return nil, nil // 未命中缓存，返回 nil
	}

	code := &dto.Code{}
	if err = json.Unmarshal([]byte(data), code); err != nil {
		fmt.Println("json.Unmarshal() err: ", err)
		return nil, err
	}
	fmt.Printf("成功命中redis缓存: %#v\n", code)

	// 保存到 MySQL，确保一致性
	entityCode := entity.Code{}.DtoToCode(code)
	if _, err = s.repo.SaveCode(entityCode); err != nil {
		fmt.Println("s.repo.SaveCode() ", err)
		return nil, err
	}

	return code, nil
}

// checkMySQL 检查 MySQL 记录并同步到 Redis
func (s *codeDomainService) checkMySQL(key string) (*dto.Code, error) {
	entityCode, ok, err := s.repo.GetCodeByHash(key)
	if err != nil {
		fmt.Println("repo.GetCodeByHash() ", err)
		return nil, err
	}
	if !ok {
		return nil, nil // 未命中记录，返回 nil
	}

	code := entityCode.CodeToDto()
	fmt.Printf("成功命中mysql记录: %#v\n", code)

	// 同步到 Redis，保证数据一致性
	if err = s.SaveToRedis(key, code); err != nil {
		fmt.Println("s.SaveToRedis() ", err)
		return nil, err
	}

	return code, nil
}

// FetchAndSave 从 LLM 获取数据并保存到 MySQL、Redis、布隆过滤器
func (s *codeDomainService) FetchAndSave(req *dto.CodeReq, key string) (*dto.Code, error) {
	s.bf.Add([]byte(key))
	fmt.Println("成功将记录缓存到布隆过滤器:", []byte(key))

	streamChan1, streamChan2, err := utils.GenerateStream(s.sign.GetCodeFlag(), req)
	if err != nil {
		fmt.Println("utils.GenerateStream() err: ", err)
		return nil, err
	}

	dtoCode := &dto.Code{Stream: streamChan1}

	go func() {
		var completeResponse strings.Builder
		for chunk := range streamChan2 {
			completeResponse.WriteString(chunk)
		}

		totalStr := completeResponse.String()
		code := &entity.Code{
			Key:         key,
			Explanation: totalStr,
			Question:    req.Question,
		}

		code.ID, err = s.repo.SaveCode(code)
		if err != nil {
			fmt.Println("s.repo.SaveCode() err: ", err)
			return
		}

		dtoCode = code.CodeToDto()
		err = s.SaveToRedis(key, dtoCode)
		if err != nil {
			fmt.Println("s.SaveToRedis() err: ", err)
			return
		}

		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			fmt.Println("s.repo.SaveHistory() ", err)
			return
		}
	}()

	return dtoCode, nil
}

func (s *codeDomainService) SaveToRedis(key string, code *dto.Code) (err error) {
	data, err := json.Marshal(code)
	if err != nil {
		fmt.Println("json.Marshal() err: ", err)
		return
	}

	if err = s.redisClient.Set(key, string(data), 24*time.Hour); err != nil {
		fmt.Println("s.redisClient.Set() err: ", err)
		return
	}

	fmt.Printf("成功将记录缓存到redis: %s —— %#v\n", key, code)
	return
}
