package service

import "siwuai/internal/domain/model/dto"

type TokenDomainService interface {
	GenerateToken(req *dto.TokenReq, secretKey string, generateTokenKey string) (string, error)
}
