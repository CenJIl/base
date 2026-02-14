package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Address  string `toml:"address"` // Redis 地址
	Password string `toml:"password"` // Redis 密码
	DB       int    `toml:"db"`       // Redis DB
}

// Client Redis 客户端（全局使用）
var Client *redis.Client

// InitRedis 初始化 Redis
//
// 使用方式：
//   if err := web.InitRedis(cfg.RedisConfig); err != nil {
//       logger.Errorf("Failed to init redis: %v", err)
//   }
func InitRedis(cfg RedisConfig) error {
	if cfg.Address == "" {
		return nil // 未配置，跳过
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 100,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	return nil
}

// Get 获取缓存
//
// 使用方式：
//   val, err := web.Get(ctx, "user:123").Result()
func Get(ctx context.Context, key string) *redis.StringCmd {
	return Client.Get(ctx, key)
}

// Set 设置缓存
//
// 使用方式：
//   err := web.Set(ctx, "user:123", "data", 10*time.Minute).Err()
func Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return Client.Set(ctx, key, value, expiration)
}

// Del 删除缓存
//
// 使用方式：
//   err := web.Del(ctx, "user:123").Err()
func Del(ctx context.Context, key string) *redis.IntCmd {
	return Client.Del(ctx, key)
}

// Close 关闭 Redis 连接
//
// 使用方式：
//   defer cache.Close()
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
