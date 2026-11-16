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
	if err := migrateUploadURLs(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrateUploadURLs(db *sql.DB) error {
	statements := []string{
		`UPDATE site_config SET value='/api/uploads/' || substr(value, 10) WHERE value LIKE '/uploads/%'`,
		`UPDATE articles SET cover='/api/uploads/' || substr(cover, 10) WHERE cover LIKE '/uploads/%'`,
		`UPDATE friend_links SET logo='/api/uploads/' || substr(logo, 10) WHERE logo LIKE '/uploads/%'`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
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

CREATE TABLE IF NOT EXISTS friend_links (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL,
	url         TEXT NOT NULL,
	logo        TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	group_name  TEXT NOT NULL DEFAULT '',
	sort_order  INTEGER NOT NULL DEFAULT 0,
	visible     INTEGER NOT NULL DEFAULT 1,
	created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_friend_links_visible_sort ON friend_links(visible,sort_order,id);

CREATE TABLE IF NOT EXISTS sponsors (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	name         TEXT NOT NULL DEFAULT '',
	amount       TEXT NOT NULL,
	currency     TEXT NOT NULL DEFAULT 'CNY',
	message      TEXT NOT NULL DEFAULT '',
	url          TEXT NOT NULL DEFAULT '',
	anonymous    INTEGER NOT NULL DEFAULT 0,
	visible      INTEGER NOT NULL DEFAULT 1,
	sponsored_at TEXT NOT NULL,
	created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sponsors_visible_date ON sponsors(visible,sponsored_at,id);

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
	freshDatabase := c == 0
	if freshDatabase {
		db.Exec(`INSERT INTO categories (name, slug) VALUES ('炼丹笔记','tech'),('洞府日常','life'),('悟道心得','essay')`)
	}

	defaults := map[string]string{
		"title":             "ZhangPanda",
		"subtitle":          "AI修道院 · 科技修仙者",
		"tagline":           "以最少的灵材，炼最完整的法宝",
		"description":       "道号熊猫道君，渡劫期修士，主修《AI 驭算诀》",
		"author":            "熊猫道君",
		"github":            "https://github.com/zhangpanda",
		"email":             "hi@zhangpanda.com",
		"sponsor_markdown":  "如果本站内容对你有所帮助，欢迎请熊猫道君喝杯灵茶。每一份支持都会用于站点维护和持续创作。",
		"sponsor_wechat_qr": "",
		"sponsor_alipay_qr": "",
		"about_markdown": `## ⚔️ 修仙简历

**道号：** 熊猫道君

**宗门：** AI修道院

**境界：** 渡劫期（MCP / 开源）

**道龄：** 十五载

**主修功法：** 《AI 驭算诀》

**道统铭：** 以算力为灵气，以数据为丹砂；以模型悟大道，以智能证长生。

---

## 🔥 修炼履历

| 境界 | 对应 | 成就 |
|------|------|------|
| 炼气 | Hello World / HTML | 踏入修仙之门 |
| 筑基 | PHP + MySQL CRUD | 初窥数据之道 |
| 金丹 | Laravel / ThinkPHP 全栈 | 凝练第一枚金丹 |
| 元婴 | Go 重构 + 微服务 | 元婴出窍，一念化万法 |
| 化神 | Elasticsearch + Kafka + 数据引擎 | 执掌数据洪流 |
| 炼虚 | Docker / K8S / CI/CD | 万象阵法，虚实一体 |
| 合体 | NLP 服务集群 + 舆情系统 | 人机合一 |
| 大乘 | LLM 应用 + Agent 开发 | 大道初成 |
| 渡劫 | MCP 框架 / 开源项目上线 | 天劫已渡 |
| 飞升 | AGI | 未至 ∞ |

---

## 🗡️ 法宝

- 🔌 **GoMCP** —— 通灵法阵（MCP 服务器框架）
- 🛒 **GoShop** —— 万商令（完整电商系统）
- 🐼 **Panda Site** —— 洞府门户（此站点）
- 🎬 **Story Flicks** —— 幻影术（AI 视频生成）

---

## ⚡ 灵根属性

- **后端**：Go / Python / PHP
- **数据库**：MySQL / SQLite / Elasticsearch / Redis / RabbitMQ / Kafka
- **阵法**：Docker / Kubernetes / Nginx / CI/CD
- **前端**：Vue / React / 微信小程序

---

## 📡 联络玉简

- GitHub: [zhangpanda](https://github.com/zhangpanda)
- 邮箱: hi@zhangpanda.com`,
	}
	for k, v := range defaults {
		db.Exec(`INSERT OR IGNORE INTO site_config (key,value) VALUES (?,?)`, k, v)
	}

	if !freshDatabase {
		return
	}

	projects := []struct {
		name, desc, url, github, icon, tags, group string
		sort                                       int
	}{
		{"GoMCP", "通灵法阵 —— Go 语言 MCP 服务器框架，类似 Gin 的开发体验", "", "https://github.com/zhangpanda/gomcp", "🔌", "Go,MCP,AI", "本命法器", 1},
		{"GoShop", "万商令 —— Go 实现的开源电商系统，392 条路由，完整电商能力", "", "https://github.com/zhangpanda/goshop", "🛒", "Go,电商", "本命法器", 2},
		{"Panda Site", "洞府门户 —— Go + SQLite 驱动的个人站点，零依赖部署", "https://zhangpanda.com", "https://github.com/zhangpanda/panda-site", "🐼", "Go,SQLite,博客", "炼器作品", 3},
		{"村务公开系统", "山河社稷图 —— 14 个功能模块，单机部署开箱即用", "", "https://github.com/zhangpanda/village-admin", "🏘️", "Go,SQLite", "炼器作品", 4},
		{"Story Flicks", "幻影术 —— AI 故事视频生成器，输入主题自动生成视频", "", "https://github.com/zhangpanda/story-flicks", "🎬", "Python,AI", "炼器作品", 5},
	}
	for _, p := range projects {
		db.Exec(`INSERT INTO projects (name,description,url,github,icon,tags,group_name,sort_order,visible) VALUES (?,?,?,?,?,?,?,?,1)`,
			p.name, p.desc, p.url, p.github, p.icon, p.tags, p.group, p.sort)
	}
}
