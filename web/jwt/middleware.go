package jwt

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	jwtMiddleware "github.com/hertz-contrib/jwt"
)

var (
	authMiddleware *jwtMiddleware.HertzJWTMiddleware
	cfg            Config
	initialized    bool
)

func Init(config Config) error {
	if config.Secret == "" {
		return ErrSecretRequired
	}

	timeout := time.Duration(config.Timeout) * time.Second
	maxRefresh := time.Duration(config.MaxRefresh) * time.Second

	var err error
	authMiddleware, err = jwtMiddleware.New(&jwtMiddleware.HertzJWTMiddleware{
		Realm:         config.Realm,
		Key:           []byte(config.Secret),
		Timeout:       timeout,
		MaxRefresh:    maxRefresh,
		IdentityKey:   config.IdentityKey,
		TokenLookup:   config.TokenLookup,
		TokenHeadName: "Bearer",
		SendCookie:    true,
		CookieName:    "token",
		CookieMaxAge:  timeout,
	})

	if err != nil {
		return err
	}

	cfg = config
	initialized = true
	return nil
}

func Middleware() app.HandlerFunc {
	if !initialized {
		return func(ctx context.Context, c *app.RequestContext) {
			c.Next(ctx)
		}
	}
	return authMiddleware.MiddlewareFunc()
}

func LoginHandler() app.HandlerFunc {
	if !initialized {
		return nil
	}
	return authMiddleware.LoginHandler
}

func GetUserID(c *app.RequestContext) string {
	if !initialized {
		return ""
	}
	claims := jwtMiddleware.ExtractClaims(context.Background(), c)
	if id, ok := claims[cfg.IdentityKey].(string); ok {
		return id
	}
	return ""
}

func GetClaims(c *app.RequestContext) map[string]interface{} {
	if !initialized {
		return nil
	}
	return jwtMiddleware.ExtractClaims(context.Background(), c)
}

func IsEnabled() bool {
	return initialized
}

func GetConfig() Config {
	return cfg
}

var ErrSecretRequired = &JWTError{Message: "JWT secret is required"}

type JWTError struct {
	Message string
}

func (e *JWTError) Error() string {
	return e.Message
}
