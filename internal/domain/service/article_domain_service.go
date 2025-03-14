package service

import (
	"siwuai/internal/domain/model/dto"
)

type ArticleDomainServiceInterface interface {
	VerifyHash(key string) (*dto.Article, error)
}
