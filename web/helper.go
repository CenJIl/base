package web

import (
	"github.com/CenJIl/base/web/cache"
	"github.com/CenJIl/base/web/database"
)

// InitDB 初始化数据库（便捷函数）
//
// 使用方式：
//
//	if err := web.InitDB(config.Database); err != nil {
//	    logger.Errorf("Failed to init database: %v", err)
//	}
func InitDB(cfg DatabaseConfig) error {
	return database.InitDB(cfg)
}

// InitRedis 初始化 Redis（便捷函数）
//
// 使用方式：
//
//	if err := web.InitRedis(config.Redis); err != nil {
//	    logger.Errorf("Failed to init redis: %v", err)
//	}
func InitRedis(cfg RedisConfig) error {
	return cache.InitRedis(cfg)
}
