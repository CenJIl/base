package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

// Transaction 事务辅助函数
//
// 使用方式：
//
//	result, err := database.Transaction(ctx, c, func(tx *sql.Tx) (any, error) {
//	    _, err := tx.Exec("UPDATE users SET name = ? WHERE id = ?", "newname", 123)
//	    return nil, err
//	})
func Transaction(ctx context.Context, c *app.RequestContext, fn func(*sql.Tx) (any, error)) (any, error) {
	// 从上下文获取事务
	tx, ok := c.MustGet("tx").(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("no transaction in context")
	}

	// 执行业务逻辑
	result, err := fn(tx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DBMiddleware 数据库事务中间件
//
// 每个请求自动开启事务，提交或回滚
//
// 使用方式：
//
//	h.Use(database.DBMiddleware())
func DBMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if DB == nil {
			c.Next(ctx)
			return
		}

		// 开启事务
		tx, err := DB.Begin()
		if err != nil {
			logger.Errorf("[DB] Failed to begin transaction: %v", err)
			c.Set("tx_error", err)
			c.Next(ctx)
			return
		}

		// 存储到上下文
		c.Set("tx", tx)

		// 处理请求
		c.Next(ctx)

		// 检查是否有错误，决定提交或回滚
		if err, ok := c.Get("tx_error"); ok && err != nil {
			logger.Warnf("[DB] Rolling back transaction due to error: %v", err)
			tx.Rollback()
		} else {
			logger.Debug("[DB] Committing transaction")
			tx.Commit()
		}
	}
}

// GetTx 从上下文获取事务
//
// 使用方式：
//
//	tx := database.GetTx(c)
func GetTx(c *app.RequestContext) *sql.Tx {
	tx, _ := c.Get("tx")
	return tx.(*sql.Tx)
}
