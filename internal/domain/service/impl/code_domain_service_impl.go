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
const lockTTL = 100 * time.Second

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
		err = fmt.Errorf("s.GetAnswer() %v", err)
		return
	}

	if code.ID != 0 {
		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			err = fmt.Errorf("s.repo.SaveHistory() %v", err)
			return
		}
	}

	return
}

// GetAnswer 用于得到代码解释信息
func (s *codeDomainService) GetAnswer(req *dto.CodeReq) (code *dto.Code, err error) {
	key, err := utils.Hash(req.Question)
	if err != nil {
		err = fmt.Errorf("utils.Hash() %v", err)
		return
	}

	// 尝试设置锁，locked为true表示设置锁成功
	locked, err := s.redisClient.TryLock(key, lockTTL)
	if err != nil {
		err = fmt.Errorf("TryLock() %v", err)
		return
	}
	fmt.Println("Locked: ", locked)

	if locked {
		// 设置了锁，别的进程此时无法访问以下资源

		// 1. 检查布隆过滤器
		if !s.bf.Test([]byte(key)) {
			fmt.Println("未命中布隆过滤器", key)
			return s.FetchAndSave(req, key)
		}
		fmt.Println("成功命中布隆过滤器，开始查询缓存...")

		// 2. 检查 Redis 缓存
		if code, err = s.checkRedis(key); err == nil && code != nil {
			s.redisClient.Unlock(key)
			return code, nil
		}
		fmt.Printf("未命中redis缓存: %s\n", key)

		// 3. 检查 MySQL 记录
		if code, err = s.checkMySQL(key); err == nil && code != nil {
			s.redisClient.Unlock(key)
			return code, nil
		}
		fmt.Printf("未命中mysql记录: %s\n", key)

		// 4. 若布隆过滤器命中，但 Redis 和 MySQL 中都未查到，则调用 LLM
		return s.FetchAndSave(req, key)
	} else {
		// 未获取锁，表示该锁正在被别人占用，等待并查询缓存
		count := 1
		for i := 1; i < 120; i++ {
			fmt.Printf("第%d次循环查询缓存:", count)
			if code, err = s.checkRedis(key); err == nil && code != nil {
				return code, nil
			}
			if code, err = s.checkMySQL(key); err == nil && code != nil {
				return code, nil
			}

			time.Sleep(5 * time.Second)
			count++
		}
		err = fmt.Errorf("轮询进行redis、mysql查询时错误：%v", err)
		fmt.Println("轮询超时")
		return nil, err
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
		err = fmt.Errorf("json.Unmarshal() err: %v", err)
		return nil, err
	}
	fmt.Printf("成功命中redis缓存: %#v\n", code)

	// 保存到 MySQL，确保一致性
	entityCode := entity.Code{}.DtoToCode(code)
	if _, err = s.repo.SaveCode(entityCode); err != nil {

		err = fmt.Errorf("s.repo.SaveCode() %v", err)
		return nil, err
	}

	return code, nil
}

// checkMySQL 检查 MySQL 记录并同步到 Redis
func (s *codeDomainService) checkMySQL(key string) (*dto.Code, error) {
	entityCode, ok, err := s.repo.GetCodeByHash(key)
	if err != nil {
		err = fmt.Errorf("repo.GetCodeByHash() %v", err)
		return nil, err
	}
	if !ok {
		return nil, nil // 未命中记录，返回 nil
	}

	code := entityCode.CodeToDto()
	fmt.Printf("成功命中mysql记录: %#v\n", code)

	// 同步到 Redis，保证数据一致性
	if err = s.SaveToRedis(key, code); err != nil {
		err = fmt.Errorf("s.SaveToRedis() %v", err)
		return nil, err
	}

	return code, nil
}

// FetchAndSave 从 LLM 获取数据并保存到 MySQL、Redis、布隆过滤器
func (s *codeDomainService) FetchAndSave(req *dto.CodeReq, key string) (*dto.Code, error) {
	streamChan1, streamChan2, err := utils.GenerateStream(s.sign.GetCodeFlag(), req)
	if err != nil {
		err = fmt.Errorf("utils.GenerateStream() %v", err)
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

		// 先添加到布隆过滤器
		s.bf.Add([]byte(key))
		fmt.Println("成功将记录缓存到布隆过滤器:", []byte(key))

		code.ID, err = s.repo.SaveCode(code)
		if err != nil {
			err = fmt.Errorf("s.repo.SaveCode() %v", err)
			return
		}

		dtoCode = code.CodeToDto()
		err = s.SaveToRedis(key, dtoCode)
		if err != nil {
			err = fmt.Errorf("s.SaveToRedis() %v", err)
			return
		}

		err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
		if err != nil {
			err = fmt.Errorf("s.repo.SaveHistory() %v", err)
			return
		}

		err = s.redisClient.Unlock(key)
		if err != nil {
			err = fmt.Errorf("s.redisClient.Unlock() %v", err)
			return
		}

	}()

	return dtoCode, nil
}

func (s *codeDomainService) SaveToRedis(key string, code *dto.Code) (err error) {
	data, err := json.Marshal(code)
	if err != nil {
		err = fmt.Errorf("json.Marshal() err: %v", err)
		return
	}

	if err = s.redisClient.Set(key, string(data), 24*time.Hour); err != nil {
		err = fmt.Errorf("s.redisClient.Set() %v", err)
		return
	}

	fmt.Printf("成功将记录缓存到redis: %s —— %#v\n", key, code)
	return
}
