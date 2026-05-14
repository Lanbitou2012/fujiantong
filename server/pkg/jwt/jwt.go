package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID          uint   `json:"user_id"`
	Role            int8   `json:"role"`
	RoleName        string `json:"role_name"`
	AuthorID        uint64 `json:"author_id"`
	AuthorizerAppID string `json:"authorizer_appid"`
	jwt.RegisteredClaims
}

type TokenOptions struct {
	UserID          uint
	Role            int8
	RoleName        string
	AuthorID        uint64
	AuthorizerAppID string
	TTL             time.Duration
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID uint, role int8) (string, error) {
	return GenerateTokenWithOptions(TokenOptions{
		UserID: userID,
		Role:   role,
		TTL:    7 * 24 * time.Hour,
	})
}

func GenerateTokenWithOptions(options TokenOptions) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_fallback_secret" // 兜底密钥，生产环境必须配置
	}
	if options.TTL <= 0 {
		options.TTL = 7 * 24 * time.Hour
	}

	claims := CustomClaims{
		UserID:          options.UserID,
		Role:            options.Role,
		RoleName:        options.RoleName,
		AuthorID:        options.AuthorID,
		AuthorizerAppID: options.AuthorizerAppID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(options.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fujiantong",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并校验 JWT 令牌
func ParseToken(tokenString string) (*CustomClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_fallback_secret"
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
