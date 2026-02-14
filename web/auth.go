package web

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig JWT configuration
type JWTConfig struct {
	SecretKey      string        // JWT signature key
	ExpirationTime time.Duration // Token expiration time
	Issuer         string        // Token issuer
}

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var (
	jwtConfig *JWTConfig
)

// InitJWT initializes JWT with configuration
func InitJWT(secretKey string, expiration time.Duration, issuer string) {
	jwtConfig = &JWTConfig{
		SecretKey:      secretKey,
		ExpirationTime: expiration,
		Issuer:         issuer,
	}
	logger.Infof("JWT initialized: expiration=%s, issuer=%s", expiration, issuer)
}

// GenerateToken generates a JWT token
func GenerateToken(userId, username, role string) (string, error) {
	if jwtConfig == nil {
		return "", fmt.Errorf("JWT not initialized")
	}

	now := time.Now()
	claims := Claims{
		UserID:   userId,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtConfig.ExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore:  jwt.NewNumericDate(now),
			Issuer:     jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.SecretKey))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (*Claims, error) {
	if jwtConfig == nil {
		return nil, fmt.Errorf("JWT not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// JWTAuthMiddleware creates JWT authentication middleware
func JWTAuthMiddleware(skipPaths ...string) app.HandlerFunc {
	skipPathMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipPathMap[path] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())

		if skipPathMap[path] {
			c.Next(ctx)
			return
		}

		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(consts.StatusUnauthorized, map[string]any{
				"code": 401,
				"message": "Missing authorization header",
				"data": nil,
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(consts.StatusUnauthorized, map[string]any{
				"code": 401,
				"message": "Invalid authorization format",
				"data": nil,
			})
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			logger.Errorf("JWT validation failed: %v", err)
			c.JSON(consts.StatusUnauthorized, map[string]any{
				"code": 401,
				"message": "Invalid or expired token",
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next(ctx)
	}
}

// GetUserID gets user ID from context
func GetUserID(c *app.RequestContext) string {
	if userId, exists := c.Get("userId"); exists {
		return userId.(string)
	}
	return ""
}

// GetUsername gets username from context
func GetUsername(c *app.RequestContext) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

// GetRole gets user role from context
func GetRole(c *app.RequestContext) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}
