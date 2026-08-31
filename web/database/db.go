package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `toml:"driver"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
	MaxOpen  int    `toml:"maxOpen"`
	MaxIdle  int    `toml:"maxIdle"`
	SSLMode  string `toml:"sslMode"`
}

// DB 数据库连接池（供 sqlc 生成的代码使用）
var DB *sql.DB

// Drivers 支持的数据库驱动
const (
	DriverMySQL      = "mysql"
	DriverPostgreSQL = "postgres"
)

// InitDB 初始化数据库连接池（供 sqlc 使用）
//
// 使用方式：
//
//	if err := web.InitDB(config.Database); err != nil {
//	    logger.Errorf("Failed to init database: %v", err)
//	}
func InitDB(cfg DatabaseConfig) error {
	if cfg.Driver == "" {
		return nil // 未配置，跳过
	}

	dsn := buildDSN(cfg)
	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour) // 连接最大生存时间1小时

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg DatabaseConfig) string {
	switch cfg.Driver {
	case DriverMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	case DriverPostgreSQL:
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "require"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s timezone=UTC", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, sslMode)
	default:
		return ""
	}
}

// Close 关闭数据库连接
//
// 使用方式：
//
//	defer database.Close()
func Close() error {
	if DB == nil {
		return nil
	}
	db := DB
	DB = nil
	return db.Close()
}
