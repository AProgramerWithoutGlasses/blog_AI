package impl

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/service"
	"time"
)

// 定义常见的错误类型
var (
	ErrInvalidToken      = errors.New("token is invalid or malformed")
	ErrUnsupportedMethod = errors.New("unsupported signing method")
	ErrValidationFailed  = errors.New("token validation failed")
)

// TokenDomainService 定义 token 领域服务
type tokenDomainService struct {
	// 可以添加其他依赖，例如配置或存储
}

// NewTokenDomainService 创建 TokenDomainService 实例
func NewTokenDomainService() service.TokenDomainService {
	return &tokenDomainService{}
}

// ValidateToken 验证 JWT token 的合法性，不提取 claims
func (s *tokenDomainService) ValidateToken(tokenString string, secretKey string) error {
	// 解析 token，不绑定具体 claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnsupportedMethod, token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	// 处理解析错误
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return fmt.Errorf("%w: malformed token", ErrInvalidToken)
			}
			if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return fmt.Errorf("%w: token is expired or not yet valid", ErrValidationFailed)
			}
			return fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// 检查 token 是否有效
	if !token.Valid {
		return ErrInvalidToken
	}

	// token 验证通过，无需提取 claims
	return nil
}

// GenerateToken 生成token
func (s *tokenDomainService) GenerateToken(req *dto.TokenReq, secretKey string, generateTokenKey string) (tokenString string, err error) {
	// 简单验证（示例）
	if req.GenerateTokenKey != generateTokenKey {
		return "", fmt.Errorf("驳回，生成token密钥错误")
	}

	// 假设验证通过，生成 JWT
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 72).Unix(), // 72 小时有效期
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(secretKey))
	if err != nil {
		return
	}

	return
}
