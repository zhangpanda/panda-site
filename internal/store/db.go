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
		"title":       "Panda Site",
		"subtitle":    "个人站点",
		"description": "Go + SQLite 驱动的个人站点系统",
		"author":      "Panda",
	}
	for k, v := range defaults {
		db.Exec(`INSERT OR IGNORE INTO site_config (key,value) VALUES (?,?)`, k, v)
	}
}
