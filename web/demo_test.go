package web_test

import (
	"context"
	"fmt"
	"time"

	"github.com/CenJIl/base/cfg"
	"github.com/CenJIl/base/web"
	"github.com/CenJIl/base/web/cache"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============================================================================
// 场景1：文件上传下载
// ============================================================================

// ExampleScenario1_FileUpload 文件上传场景
//
// 场景：用户上传 Excel 文件导入数据
func ExampleScenario1_FileUpload() {
	cfg := web.DatabaseConfig{
		Driver: "mysql",
		Host:   "localhost",
		Port:   3306,
		User:   "root",
		DBName: "myapp",
	}

	_ = cfg // 配置示例

	_ = func(ctx context.Context, c *app.RequestContext) {
		file, _ := c.FormFile("file")
		if err := web.ValidateFile(file, web.UploadConfig{
			MaxFileSize: 10 * 1024 * 1024,
			AllowedExts: []string{".xlsx", ".xls"},
		}); err != nil {
			panic(web.BadRequestHTTP(err.Error()))
		}

		filename := web.GenerateFilename(file.Filename)
		savePath := fmt.Sprintf("./uploads/%s", filename)
		web.SaveUploadedFile(file, savePath)

		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"filename": filename,
			"url":      "/uploads/" + filename,
		}))
	}
}

// ExampleScenario2_FileDownload 文件下载场景
//
// 场景：导出数据为 Excel 提供下载
func ExampleScenario2_FileDownload() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		filename := "report.xlsx"
		filePath := fmt.Sprintf("./uploads/%s", filename)

		if !web.FileExists(filePath) {
			panic(web.NotFoundHTTP("文件不存在"))
		}

		web.DownloadWithRange(c, filePath, filename)
	}
}

// ============================================================================
// 场景2：数据库操作
// ============================================================================

// ExampleScenario3_DatabaseCRUD 数据库 CRUD 场景
//
// 场景：用户管理（创建、查询、更新、删除）
// 注意：本示例展示 sqlc 生成的代码如何使用
func ExampleScenario3_DatabaseCRUD() {
	// 前置条件：
	// 1. 运行 sqlc generate 生成 db/ 目录下的代码
	// 2. 生成的代码包括：db.go, models.go, user.sql.go 等

	// sqlc 生成的函数（示例）：
	// - CreateUser(ctx, db, name, email, password, role) error
	// - GetUserByID(ctx, db, userID) (*User, error)
	// - ListUsers(ctx, db, limit, offset) ([]User, error)
	// - UpdateUser(ctx, db, userUpdateParams) error
	// - DeleteUser(ctx, db, userID) error

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 创建用户（使用 sqlc 生成的函数）
		// err := db.CreateUser(ctx, database.DB, "张三", "zhangsan@example.com", "hashed_password", "user")
		// if err != nil {
		//     panic(web.InternalHTTP("创建用户失败"))
		// }

		c.JSON(consts.StatusOK, web.Success(nil))
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 查询用户（使用 sqlc 生成的函数）
		// user, err := db.GetUserByID(ctx, database.DB, 123)
		// if err != nil {
		//     panic(web.NotFoundHTTP("用户不存在"))
		// }

		// c.JSON(consts.StatusOK, web.Success(user))
		_ = ctx
		_ = c
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 更新用户（使用 sqlc 生成的函数）
		// err := db.UpdateUser(ctx, database.DB, db.UpdateUserParams{
		//     ID: 123,
		//     Name: "新名字",
		//     Email: "newemail@example.com",
		// })
		// if err != nil {
		//     panic(web.InternalHTTP("更新用户失败"))
		// }

		c.JSON(consts.StatusOK, web.Success(nil))
		_ = ctx
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 删除用户（使用 sqlc 生成的函数）
		// userId, _ := strconv.Atoi(c.Param("id"))
		// err := db.DeleteUser(ctx, database.DB, int64(userId))
		// if err != nil {
		//     panic(web.InternalHTTP("删除用户失败"))
		// }

		c.JSON(consts.StatusOK, web.Success(nil))
		_ = ctx
	}
}

// ============================================================================
// 场景3：Redis 缓存
// ============================================================================

// ExampleScenario4_RedisCache Redis 缓存场景
//
// 场景：热点数据缓存、Session 管理
func ExampleScenario4_RedisCache() {
	cfg := web.RedisConfig{
		Address: "localhost:6379",
	}

	_ = cfg

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 设置缓存
		key := "user:123"
		err := cache.Set(ctx, key, map[string]any{
			"id":       123,
			"username": "admin",
		}, 10*time.Minute).Err()
		if err != nil {
			panic(web.InternalHTTP("设置缓存失败"))
		}

		c.JSON(consts.StatusOK, web.Success(nil))
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 获取缓存
		key := "user:123"
		val, err := cache.Get(ctx, key).Result()
		if err != nil {
			panic(web.InternalHTTP("获取缓存失败"))
		}

		c.JSON(consts.StatusOK, web.Success(val))
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 删除缓存
		key := "user:123"
		err := cache.Del(ctx, key).Err()
		if err != nil {
			panic(web.InternalHTTP("删除缓存失败"))
		}

		c.JSON(consts.StatusOK, web.Success(nil))
	}
}

// ============================================================================
// 场景4：API 错误处理
// ============================================================================

// ExampleScenario5_APIErrorHandling API 错误处理场景
//
// 场景：RESTful API 的统一错误处理
func ExampleScenario5_APIErrorHandling() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 参数验证失败
		email := c.PostForm("email")
		if email == "" {
			panic(web.BadRequestHTTP("邮箱不能为空"))
		}
		// 自动被 ExceptionHandler 捕获并转换为：
		// {"code": 400, "message": "邮箱不能为空", "data": null}
		c.JSON(consts.StatusOK, web.Success(nil))
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 资源不存在
		userId := c.Param("id")
		if !userExists(userId) {
			panic(web.NotFoundHTTP("用户不存在"))
		}
		// 自动被 ExceptionHandler 捕获并转换为：
		// {"code": 404, "message": "用户不存在", "data": null}
		c.JSON(consts.StatusOK, web.Success(nil))
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 业务逻辑错误
		userId := c.Param("id")
		if hasUnfinishedOrders(userId) {
			panic(web.NewException(20001, "用户有未完成订单，无法删除"))
		}
		// 自动被 ExceptionHandler 捕获并转换为：
		// {"code": 20001, "message": "用户有未完成订单，无法删除", "data": null}
		c.JSON(consts.StatusOK, web.Success(nil))
	}
}

// ============================================================================
// 场景5：统一响应格式
// ============================================================================

// ExampleScenario6_UnifiedResponse 统一响应格式场景
//
// 场景：所有接口返回统一的 JSON 格式
func ExampleScenario6_UnifiedResponse() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 成功响应（无数据）
		c.JSON(consts.StatusOK, web.Success(nil))
		// 输出：{"code": 0, "message": "success", "data": null, "traceId": "..."}
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 成功响应（有数据）
		user := map[string]string{
			"id":       "123",
			"username": "admin",
		}
		c.JSON(consts.StatusOK, web.Success(user))
		// 输出：{"code": 0, "message": "success", "data": {"id":"123","username":"admin"}, "traceId": "..."}
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 分页响应
		users := []map[string]string{}
		c.JSON(consts.StatusOK, web.PagedSuccess(users, 1, 20, 100))
		// 输出：{"code": 0, "message": "success", "data": {"items":[...], "page":1, "pageSize":20, "total":100, "totalPage":5}, "traceId": "..."}
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 失败响应（带数据）
		c.JSON(consts.StatusOK, web.FailWithData(400, "账号或密码错误", map[string]string{
			"remainingAttempts": "3",
		}))
		// 输出：{"code": 400, "message": "账号或密码错误", "data": {"remainingAttempts":"3"}, "traceId": "..."}
	}
}

// ============================================================================
// 场景6：JWT 认证
// ============================================================================

// ExampleScenario7_JWTAuth JWT 认证场景
//
// 场景：需要登录的 API 接口
func ExampleScenario7_JWTAuth() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 登录接口（无需认证）
		username := c.PostForm("username")
		password := c.PostForm("password")

		// 验证用户（假设用户存在且密码正确）
		if username == "admin" && password == "admin" {
			token, _ := web.GenerateToken("1", username, "admin")

			c.JSON(consts.StatusOK, web.Success(map[string]string{
				"token": token,
			}))
		} else {
			panic(web.UnauthorizedHTTP("账号或密码错误"))
		}
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 需要认证的接口（从上下文获取用户信息）
		// 在 JWT 中间件中已设置：c.Set("user_id", userId)
		userId := web.GetUserID(c)
		username := web.GetUsername(c)

		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"user_id":  userId,
			"username": username,
		}))
	}
}

// ============================================================================
// 场景7：限流
// ============================================================================

// ExampleScenario8_RateLimit 限流场景
//
// 场景：防止 API 被恶意调用
func ExampleScenario8_RateLimit() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 短信接口限流
		c.JSON(consts.StatusOK, web.Success(nil))
	}

	// 限流中间件会在引擎层自动应用
	// 超过限制返回 429 Too Many Requests
}

// ============================================================================
// 场景8：中间件链
// ============================================================================

// ExampleScenario9_MiddlewareChain 中间件链场景
//
// 场景：多个中间件组合使用
func ExampleScenario9_MiddlewareChain() {
	cfg := web.DatabaseConfig{
		Driver: "mysql",
		Host:   "localhost",
		Port:   3306,
		User:   "root",
		DBName: "myapp",
	}

	redisCfg := web.RedisConfig{
		Address: "localhost:6379",
	}

	_ = cfg
	_ = redisCfg

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 中间件链（顺序很重要）
		// 1. 全局异常处理（已在 NewServer 中自动注册）
		// 2. 限流中间件
		// 3. JWT 认证中间件（跳过登录接口）
		// 4. 数据库事务中间件
		c.JSON(consts.StatusOK, web.Success(nil))
	}
}

// ============================================================================
// 场景9：完整项目初始化
// ============================================================================

// ExampleScenario10_CompleteProject 完整项目初始化场景
//
// 场景：从零开始搭建一个完整的 Web 项目
func ExampleScenario10_CompleteProject() {
	// 步骤1：定义配置结构
	type AppConfig struct {
		AppName    string `toml:"appName"`
		JWTSecret  string `toml:"jwtSecret"`
		web.Config        // 必须内嵌 web.Config
	}

	// 步骤2：初始化配置
	defaultConfig := []byte(`
appName = "MyApp"
debug = true
jwtSecret = "your-secret-key-change-in-production"

[web]
port = 8080
localePath = "./locales"
defaultLang = "zh-CN"
logLevel = "info"

[web.upload]
maxFileSize = 10485760        # 10MB
allowedExts = [".jpg", ".png", ".pdf", ".xlsx"]
uploadPath = "./uploads"
urlPrefix = "/uploads"

[web.database]
driver = "mysql"
host = "localhost"
port = 3306
user = "root"
password = ""
dbname = "myapp"
maxOpen = 100
maxIdle = 10

[web.redis]
address = "localhost:6379"
password = ""
db = 0
`)

	cfg.InitConfig[AppConfig](defaultConfig)
	config := cfg.GetCfg[AppConfig]()

	// 步骤3：初始化数据库
	if err := web.InitDB(config.Database); err != nil {
		panic(fmt.Sprintf("Failed to init database: %v", err))
	}

	// 步骤4：初始化 Redis
	if err := web.InitRedis(config.Redis); err != nil {
		panic(fmt.Sprintf("Failed to init redis: %v", err))
	}

	// 步骤5：创建服务器
	h := web.NewServer[AppConfig]()

	// 步骤6：注册路由
	h.POST("/login", loginHandler)
	h.GET("/profile", profileHandler)
	h.POST("/users", createUserHandler)

	// 步骤7：启动服务
	_ = h
}

// ============================================================================
// 辅助函数（仅示例用）
// ============================================================================

var loginHandler app.HandlerFunc = func(ctx context.Context, c *app.RequestContext) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "admin" && password == "admin" {
		token, _ := web.GenerateToken("1", username, "admin")
		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"token": token,
		}))
	} else {
		panic(web.UnauthorizedHTTP("账号或密码错误"))
	}
}

var profileHandler app.HandlerFunc = func(ctx context.Context, c *app.RequestContext) {
	// 这里需要验证 JWT token
	// 简化示例，直接返回用户信息
	c.JSON(consts.StatusOK, web.Success(map[string]any{
		"user_id":  1,
		"username": "admin",
		"role":     "admin",
	}))
}

var createUserHandler app.HandlerFunc = func(ctx context.Context, c *app.RequestContext) {
	// 创建用户（使用 sqlc 生成的函数）
	// name := c.PostForm("name")
	// email := c.PostForm("email")
	// password := c.PostForm("password")
	//
	// err := db.CreateUser(ctx, database.DB, name, email, hashPassword(password), "user")
	// if err != nil {
	//     panic(web.InternalHTTP("创建用户失败"))
	// }

	c.JSON(consts.StatusOK, web.Success(nil))
	_ = ctx
}

func userExists(id string) bool {
	return false
}

func hasUnfinishedOrders(id string) bool {
	return false
}
