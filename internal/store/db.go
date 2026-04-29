package store

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(schema); err != nil {
		return nil, err
	}
	seed(db)
	return db, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS categories (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS articles (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	title        TEXT NOT NULL,
	slug         TEXT NOT NULL UNIQUE,
	content      TEXT NOT NULL DEFAULT '',
	html         TEXT NOT NULL DEFAULT '',
	summary      TEXT NOT NULL DEFAULT '',
	cover        TEXT NOT NULL DEFAULT '',
	category_id  INTEGER NOT NULL DEFAULT 0,
	tags         TEXT NOT NULL DEFAULT '',
	status       TEXT NOT NULL DEFAULT 'draft',
	views        INTEGER NOT NULL DEFAULT 0,
	published_at DATETIME,
	created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);

CREATE TABLE IF NOT EXISTS projects (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	url         TEXT NOT NULL DEFAULT '',
	github      TEXT NOT NULL DEFAULT '',
	icon        TEXT NOT NULL DEFAULT '',
	tags        TEXT NOT NULL DEFAULT '',
	group_name  TEXT NOT NULL DEFAULT '',
	sort_order  INTEGER NOT NULL DEFAULT 0,
	visible     INTEGER NOT NULL DEFAULT 1,
	created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS site_config (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS admin (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL
);
`

func seed(db *sql.DB) {
	var c int
	db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&c)
	if c > 0 {
		return
	}
	db.Exec(`INSERT INTO categories (name, slug) VALUES ('技术','tech'),('生活','life'),('随笔','essay')`)

	defaults := map[string]string{
		"title":       "ZhangPanda",
		"subtitle":    "全栈开发 · 开源爱好者",
		"tagline":     "用最少的依赖，做完整的产品",
		"description": "Go / Python / PHP 全栈开发，专注开源工具和系统开发",
		"author":      "ZhangPanda",
		"github":      "https://github.com/zhangpanda",
		"email":       "hi@zhangpanda.com",
		"about_markdown": `👋 **Hi**

我是 ZhangPanda，一个有十二年开发经验的全栈开发者。主导过舆情监测、数据路由引擎、NLP 服务集群、LLM Agent 等核心系统的开发。业余时间做开源，喜欢用最少的依赖做完整的产品。

## 技术栈

- **后端**：Go / Python / PHP
- **数据库**：MySQL / SQLite / Elasticsearch / Redis / RabbitMQ / Kafka / HBase / Lindorm
- **部署**：Docker / K8S / Nginx / OpenResty / Caddy / CI/CD
- **前端**：Vue / React / uni-app / 微信小程序

## 联系我

- GitHub: [zhangpanda](https://github.com/zhangpanda)
- 邮箱: hi@zhangpanda.com`,
	}
	for k, v := range defaults {
		db.Exec(`INSERT OR IGNORE INTO site_config (key,value) VALUES (?,?)`, k, v)
	}

	projects := []struct {
		name, desc, url, github, icon, tags, group string
		sort                                       int
	}{
		{"GoMCP", "Go 语言 MCP 服务器框架，类似 Gin 的开发体验，已收录于 Glama MCP 目录", "", "https://github.com/zhangpanda/gomcp", "🔌", "Go,MCP,AI", "开源项目", 1},
		{"GoShop", "Go 实现的开源电商系统，对标 ShopXO，392 条路由，完整电商能力", "", "https://github.com/zhangpanda/goshop", "🛒", "Go,电商", "开源项目", 2},
		{"Panda Site", "Go + SQLite 驱动的个人站点系统，零依赖部署，自带后台管理", "https://zhangpanda.com", "https://github.com/zhangpanda/panda-site", "🐼", "Go,SQLite,博客", "开源项目", 3},
		{"村务公开系统", "Go + SQLite，14 个功能模块，单机部署开箱即用", "", "https://github.com/zhangpanda/village-admin", "🏘️", "Go,SQLite", "开源项目", 4},
		{"Story Flicks", "AI 故事视频生成器，输入主题自动生成含配图、语音、字幕的视频", "", "https://github.com/zhangpanda/story-flicks", "🎬", "Python,AI", "开源项目", 5},
	}
	for _, p := range projects {
		db.Exec(`INSERT INTO projects (name,description,url,github,icon,tags,group_name,sort_order,visible) VALUES (?,?,?,?,?,?,?,?,1)`,
			p.name, p.desc, p.url, p.github, p.icon, p.tags, p.group, p.sort)
	}
}
