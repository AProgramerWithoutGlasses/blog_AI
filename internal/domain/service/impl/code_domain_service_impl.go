package impl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/redis_utils"
	"siwuai/internal/infrastructure/utils"
	"siwuai/pkg/globals"
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

func (s *codeDomainService) ExplainCode(req *dto.CodeReq) (code *dto.Code, err error) {
	// 得到代码解释信息
	code, err = s.GetAnswer(req)
	if err != nil {
		fmt.Println("s.GetAnswer() ", err)
		return
	}

	// 记录 history
	err = s.repo.SaveHistory(entity.History{UserID: req.UserId, CodeID: code.ID})
	if err != nil {
		fmt.Println("s.repo.SaveHistory() ", err)
		return
	}

	return
}

// GetAnswer 用于得到代码解释信息
func (s *codeDomainService) GetAnswer(req *dto.CodeReq) (code *dto.Code, err error) {
	// 获取问题的hash值
	key, err := utils.Hash(req.Question)
	if err != nil {
		fmt.Println("utils.Hash() ", err)
		return
	}

	// 1. 检查布隆过滤器
	if !s.bf.Test([]byte(key)) {
		// 若未命中布隆过滤器
		fmt.Println("未命中布隆过滤器", key)

		// 则调用llm生成新答复，并保存记录。
		code, err = s.FetchAndSave(req, key)
		if err != nil {
			fmt.Println("s.FetchAndSave() ", err)
			return
		}

		return
	}

	// 2. 检查 Redis 缓存
	data, err := s.redisClient.Get(key)
	if err == nil && data != "" {
		// 若成功命中缓存
		code = &dto.Code{} // 初始化指针，防止空指针异常

		// 将命中结果反序列化到code并返回
		if err = json.Unmarshal([]byte(data), code); err != nil {
			fmt.Println("json.Unmarshal() err: ", err)
			return
		}
		return
	}

	// 3. 检查 MySQL 记录
	entityCode, ok, err := s.repo.GetCodeByHash(key)
	if err != nil {
		fmt.Println("repo.GetCodeByHash() ", err)
		return
	}
	// 若成功命中记录
	if ok {
		// 结构体转换
		code = entityCode.CodeToDto()

		// 同步到redis, 保证数据一致性
		err = s.SaveToRedis(key, code)
		if err != nil {
			fmt.Println("s.SaveToRedis() ", err)
			return
		}
		return
	}

	// 调用llm生成新答复，并保存记录。
	code, err = s.FetchAndSave(req, key)
	if err != nil {
		fmt.Println("s.FetchAndSave() ", err)
		return
	}

	return
}

// FetchAndSave 从 LLM 获取数据并保存到 MySQL、Redis、布隆过滤器
func (s *codeDomainService) FetchAndSave(req *dto.CodeReq, key string) (*dto.Code, error) {
	// 从 LLM 获取数据
	explain, stream, err := utils.GenerateStream(globals.CodeAICode, req)
	if err != nil {
		fmt.Println("utils.Generate() err: ", err)
		return nil, err
	}

	fmt.Println("explanation:", explain)
	fmt.Println("stream:", stream)

	// 构造dtoCode
	code := &entity.Code{
		Key:         key,
		Explanation: explain,
		Question:    req.Question,
	}
	dtoCode := code.CodeToDto()

	// 保存到 MySQL 的 code 表
	code.ID, err = s.repo.SaveCode(code)
	if err != nil {
		fmt.Println("s.repo.SaveCode() ", err)
		return nil, err
	}

	// 缓存到 Redis
	err = s.SaveToRedis(key, dtoCode)
	if err != nil {
		fmt.Println("s.repo.SaveToRedis() ", err)
		return nil, err
	}

	// 缓存到布隆过滤器
	s.bf.Add([]byte(key))
	fmt.Println("成功将记录缓存到布隆过滤器:", []byte(key))

	// 最后绑定stream，防止存入mysql、redis中
	dtoCode.Stream = stream

	return dtoCode, nil
}

func (s *codeDomainService) SaveToRedis(key string, code *dto.Code) (err error) {
	data, err := json.Marshal(code)
	if err != nil {
		fmt.Println("json.Marshal() err: ", err)
		return
	}

	if err = s.redisClient.Set(key, string(data), 24*time.Hour); err != nil { // todo redis的有效期
		fmt.Println("s.redisClient.Set() err: ", err)
		return
	}

	fmt.Printf("成功将记录缓存到redis: %s —— %s\n", key, code.Explanation)

	return
}
