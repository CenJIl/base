# Web 项目模板

此目录包含 web 模块的项目模板，用于快速创建新项目。

## 目录结构

```
_template/
├── basic/     # 基础模板（不含数据库）
└── full/      # 完整模板（含数据库 + Redis）
```

## 使用方法

```bash
# 复制模板到你的项目目录
cp -r web/_template/full /path/to/your-project

# 进入项目
cd /path/to/your-project

# 修改配置文件
vim config.toml

# 运行
go run main.go
```

## 注意

- 模板目前为空，后续逐步完善
- 模板应包含：main.go、config.example.toml、resources/locales/ 等
