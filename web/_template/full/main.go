// @title 用户管理 API
// @version 1.0
// @description 用户管理的 RESTful API 示例
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"strconv"

	"github.com/CenJIl/base/web"
	"github.com/CenJIl/base/web/jwt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type AppConfig struct {
	AppName   string `toml:"appName"`
	JWTSecret string `toml:"jwtSecret"`
	web.Config
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = map[int]User{
	1: {ID: 1, Name: "张三", Email: "zhangsan@example.com"},
	2: {ID: 2, Name: "李四", Email: "lisi@example.com"},
}

func main() {
	if err := jwt.Init(jwt.Config{
		Secret:      "your-secret-key-change-in-production",
		Realm:       "jwt",
		Timeout:     3600,
		MaxRefresh:  7200,
		IdentityKey: "identity",
		SkipPaths:   []string{"/login", "/health", "/hello"},
	}); err != nil {
		panic(err)
	}

	h := web.NewServer[AppConfig]()

	h.Use(jwt.Middleware())

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"status": "ok",
		}))
	})

	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		name := c.DefaultQuery("name", "World")
		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"message": "Hello, " + name + "!",
		}))
	})

	h.POST("/login", func(ctx context.Context, c *app.RequestContext) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "admin" && password == "admin" {
			c.JSON(consts.StatusOK, web.Success(map[string]string{
				"message": "login success",
			}))
		} else {
			panic(web.UnauthorizedHTTP("用户名或密码错误"))
		}
	})

	h.POST("/api/users", func(ctx context.Context, c *app.RequestContext) {
		name := c.PostForm("name")
		email := c.PostForm("email")

		if name == "" || email == "" {
			panic(web.BadRequestHTTP("参数不完整"))
		}

		newID := len(users) + 1
		users[newID] = User{
			ID:    newID,
			Name:  name,
			Email: email,
		}

		c.JSON(consts.StatusOK, web.Success(map[string]interface{}{
			"id":    newID,
			"name":  name,
			"email": email,
		}))
	})

	h.GET("/api/users", func(ctx context.Context, c *app.RequestContext) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

		var userList []User
		for _, u := range users {
			userList = append(userList, u)
		}

		total := int64(len(userList))
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > len(userList) {
			userList = []User{}
		} else if end > len(userList) {
			userList = userList[start:]
		} else {
			userList = userList[start:end]
		}

		c.JSON(consts.StatusOK, web.PagedSuccess(userList, page, pageSize, total))
	})

	h.GET("/api/users/:id", func(ctx context.Context, c *app.RequestContext) {
		id, _ := strconv.Atoi(c.Param("id"))
		user, ok := users[id]
		if !ok {
			panic(web.NotFoundHTTP("用户不存在"))
		}
		c.JSON(consts.StatusOK, web.Success(user))
	})

	h.PUT("/api/users/:id", func(ctx context.Context, c *app.RequestContext) {
		id, _ := strconv.Atoi(c.Param("id"))
		user, ok := users[id]
		if !ok {
			panic(web.NotFoundHTTP("用户不存在"))
		}

		name := c.PostForm("name")
		email := c.PostForm("email")

		if name != "" {
			user.Name = name
		}
		if email != "" {
			user.Email = email
		}
		users[id] = user

		c.JSON(consts.StatusOK, web.Success(user))
	})

	h.DELETE("/api/users/:id", func(ctx context.Context, c *app.RequestContext) {
		id, _ := strconv.Atoi(c.Param("id"))
		if _, ok := users[id]; !ok {
			panic(web.NotFoundHTTP("用户不存在"))
		}
		delete(users, id)
		c.JSON(consts.StatusOK, web.Success(nil))
	})

	web.MustRun[AppConfig](h)
}
