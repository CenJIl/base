# Hertz Web 项目模板

此目录提供可复制的 basic 和 full 模板。模板统一使用 `app.toml`，消费者可按需删除数据库、Redis、JWT、上传等配置与代码。

## 目录结构

```
_template/
├── basic/     # 最小模板（不含数据库、Redis）
└── full/      # 完整模板（含数据库 + Redis + i18n + JWT + Swagger）
```

## 模板说明

### basic - 最小模板

适合：学习、原型开发、微服务

包含：
- 最小配置
- 基本路由
- 无数据库依赖

### full - 完整示例

适合：需要认证、数据库和缓存示例的项目。
包含统一响应、异常处理、JWT、数据库/Redis 配置示例和用户 CRUD demo。

注意：这是演示模板，不是开箱即用的生产认证方案；必须替换 JWT Secret、账号校验和数据库凭据。

## 快速开始

```bash
# 复制模板（Windows PowerShell 可使用 Copy-Item -Recurse）
cp -r web/_template/basic /path/to/your-project
cd /path/to/your-project
go mod init your-project-name
go mod edit -replace github.com/CenJIl/base=/path/to/base
go mod tidy
go run main.go
```

启动前编辑 `app.toml`。`web.NewServer` 默认读取当前目录的 `app.toml`；数据库、Redis、i18n 和上传配置均为可选。

full 模板的 Swagger/Taskfile 仅作为可选开发示例，不代表核心库会自动启用 Swagger。

## 使用 task 管理项目

| 命令 | 说明 |
|------|------|
| `task dev` | 自动生成 Swagger 并运行 |
| `task swag` | 仅生成 Swagger 文档 |
| `task build` | 构建应用 |

## API 示例 (full 模板)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /health | 健康检查 | 否 |
| GET | /hello | Hello World | 否 |
| POST | /login | 用户登录 | 否 |
| GET | /api/users | 用户列表 | 是 |
| GET | /api/users/:id | 获取用户 | 是 |
| POST | /api/users | 创建用户 | 是 |
| PUT | /api/users/:id | 更新用户 | 是 |
| DELETE | /api/users/:id | 删除用户 | 是 |

## Swagger 文档

启动服务后访问：http://localhost:8080/swagger/index.html

## JWT 使用

```go
// 初始化 JWT
if err := jwt.Init(jwt.Config{
    Secret:    "your-secret",
    SkipPaths: []string{"/login", "/health"},
}); err != nil {
    panic(err)
}

// 注册中间件
h.Use(jwt.Middleware())

// 登录接口（无需认证）
h.POST("/login", jwt.LoginHandler(), loginHandler)

// 获取当前用户
userID := jwt.GetUserID(c)
```

## 快速对比

| 功能 | basic | full |
|------|-------|------|
| HTTP 服务 | ✅ | ✅ |
| 配置管理 | ✅ | ✅ |
| 日志 | ✅ | ✅ |
| CORS | ✅ | ✅ |
| 请求ID | ✅ | ✅ |
| 安全头 | ✅ | ✅ |
| 数据库 | ❌ | ✅ (配置) |
| Redis | ❌ | ✅ (配置) |
| i18n | ❌ | ✅ (配置) |
| JWT | ❌ | ✅ |
| CRUD 示例 | ❌ | ✅ |
| Swagger | ❌ | ✅ |
| task 自动化 | ❌ | ✅ |
