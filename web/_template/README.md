# Web 项目模板

此目录包含 web 模块的项目模板，用于快速创建新项目。

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

### full - 完整模板

适合：生产环境项目

包含：
- 完整配置
- MySQL/PostgreSQL 数据库（配置已就绪）
- Redis 缓存（配置已就绪）
- i18n 多语言支持（配置已就绪）
- JWT 认证（已配置）
- 统一错误处理
- CRUD 示例接口（用户管理）
- Swagger API 文档支持

## 快速开始

### 安装工具

```bash
# 安装 go-task (跨平台构建工具)
go install github.com/go-task/task/v3/cmd/task@latest

# 安装 swag (Swagger 文档生成工具)
go install github.com/swaggo/swag/cmd/swag@latest
```

### 启动项目

```bash
# 1. 复制模板到你的项目目录
cp -r web/_template/full /path/to/your-project

# 2. 进入项目目录
cd /path/to/your-project

# 3. 复制配置文件
cp app.toml config.toml

# 4. 初始化 Go 模块
go mod init your-project-name
go mod tidy

# 5. 运行开发服务器（自动生成 Swagger 文档）
task dev
```

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
