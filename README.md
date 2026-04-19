# 🐼 熊猫小站 (Panda Site)

Go + SQLite 驱动的个人站点系统，单二进制部署，零框架依赖。

项目展示 · 技术博客 · 个人介绍 · 管理后台，开箱即用。

## 功能

| 模块 | 说明 |
|------|------|
| 🏠 首页 | 个人介绍 + 项目卡片 + 最新文章 |
| 💼 项目展示 | 按分组展示，支持图标/标签/链接/排序/可见性 |
| 📝 博客 | Markdown 文章，分类/标签/分页/上下篇导航 |
| 👤 关于 | Markdown 可配置的个人介绍页 |
| ⚙️ 管理后台 | 文章编辑（Markdown + 实时预览）/ 项目管理 / 分类管理 / 站点设置 |
| 📡 RSS | 自动生成 RSS 订阅源 |
| 🔍 SEO | 服务端渲染，搜索引擎友好 |

## 快速开始

```bash
# 编译
go build -o panda-site ./cmd/server

# 启动
./panda-site

# 自定义端口和数据库路径
./panda-site -addr :3000 -db mysite.db
```

启动后：
- 站点：http://localhost:8080
- 管理后台：http://localhost:8080/admin/

首次访问管理后台会引导创建管理员账号。

## Docker 部署

```bash
docker build -t panda-site .
docker run -d -p 8080:8080 -v site-data:/app/data panda-site
```

## 技术栈

- Go 1.22+（net/http 标准库路由）
- SQLite（WAL 模式，单文件数据库）
- JWT + bcrypt 认证
- [Goldmark](https://github.com/yuin/goldmark) Markdown 渲染 + 代码高亮
- [Tailwind CSS](https://tailwindcss.com) CDN（无构建工具）
- 服务端渲染（Go html/template）

## 项目结构

```
panda-site/
├── cmd/server/main.go              # 入口 + 路由
├── internal/
│   ├── model/model.go              # Article, Project, Category
│   ├── store/
│   │   ├── db.go                   # Schema + 种子数据
│   │   ├── article.go              # 文章 CRUD + 上下篇
│   │   ├── project.go              # 项目 CRUD
│   │   ├── category.go             # 分类 CRUD
│   │   └── config.go               # 站点配置 KV
│   ├── handler/
│   │   ├── api.go                  # REST API + Markdown 渲染
│   │   └── page.go                 # 服务端渲染 + RSS
│   └── middleware/auth.go          # JWT 认证
└── web/
    ├── templates/                  # 前台页面模板
    │   ├── layout.html
    │   ├── index.html              # 首页
    │   ├── projects.html           # 项目展示
    │   ├── blog.html               # 博客列表
    │   ├── article.html            # 文章详情
    │   └── about.html              # 关于
    └── admin/                      # 管理后台
        ├── index.html
        └── admin.js
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| JWT_SECRET | JWT 签名密钥 | 随机生成（重启失效） |

## License

[Apache-2.0](LICENSE)
