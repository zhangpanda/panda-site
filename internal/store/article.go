package store

import (
	"database/sql"
	"panda-site/internal/model"
	"time"
)

type ArticleStore struct{ DB *sql.DB }

const artCols = `a.id,a.title,a.slug,a.summary,a.cover,a.category_id,COALESCE(c.name,''),a.tags,a.status,a.views,a.published_at,a.created_at,a.updated_at`
const artColsFull = `a.id,a.title,a.slug,a.content,a.html,a.summary,a.cover,a.category_id,COALESCE(c.name,''),a.tags,a.status,a.views,a.published_at,a.created_at,a.updated_at`
const artJoin = ` FROM articles a LEFT JOIN categories c ON a.category_id=c.id`

func scanArt(r interface{ Scan(...any) error }, a *model.Article, full bool) error {
	if full {
		return r.Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.HTML, &a.Summary, &a.Cover, &a.CategoryID, &a.Category, &a.Tags, &a.Status, &a.Views, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	}
	return r.Scan(&a.ID, &a.Title, &a.Slug, &a.Summary, &a.Cover, &a.CategoryID, &a.Category, &a.Tags, &a.Status, &a.Views, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
}

func (s *ArticleStore) List(page, size int, status, catSlug string) ([]model.Article, int) {
	where, args := " WHERE 1=1", []any{}
	if status != "" {
		where += " AND a.status=?"
		args = append(args, status)
	}
	if catSlug != "" {
		where += " AND c.slug=?"
		args = append(args, catSlug)
	}
	var total int
	s.DB.QueryRow("SELECT COUNT(*)"+artJoin+where, args...).Scan(&total)
	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query("SELECT "+artCols+artJoin+where+" ORDER BY a.id DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return []model.Article{}, 0
	}
	defer rows.Close()
	list := make([]model.Article, 0)
	for rows.Next() {
		var a model.Article
		scanArt(rows, &a, false)
		list = append(list, a)
	}
	return list, total
}

func (s *ArticleStore) GetBySlug(slug string) (*model.Article, error) {
	a := &model.Article{}
	err := scanArt(s.DB.QueryRow("SELECT "+artColsFull+artJoin+" WHERE a.slug=?", slug), a, true)
	return a, err
}

func (s *ArticleStore) GetByID(id int64) (*model.Article, error) {
	a := &model.Article{}
	err := scanArt(s.DB.QueryRow("SELECT "+artColsFull+artJoin+" WHERE a.id=?", id), a, true)
	return a, err
}

func (s *ArticleStore) Create(a *model.Article) error {
	res, err := s.DB.Exec(
		`INSERT INTO articles (title,slug,content,html,summary,cover,category_id,tags,status,published_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		a.Title, a.Slug, a.Content, a.HTML, a.Summary, a.Cover, a.CategoryID, a.Tags, a.Status, a.PublishedAt)
	if err != nil {
		return err
	}
	a.ID, _ = res.LastInsertId()
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	return nil
}

func (s *ArticleStore) Update(a *model.Article) error {
	_, err := s.DB.Exec(
		`UPDATE articles SET title=?,slug=?,content=?,html=?,summary=?,cover=?,category_id=?,tags=?,status=?,published_at=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		a.Title, a.Slug, a.Content, a.HTML, a.Summary, a.Cover, a.CategoryID, a.Tags, a.Status, a.PublishedAt, a.ID)
	return err
}

func (s *ArticleStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM articles WHERE id=?", id)
	return err
}

func (s *ArticleStore) IncrViews(id int64) {
	s.DB.Exec("UPDATE articles SET views=views+1 WHERE id=?", id)
}

func (s *ArticleStore) GetAdjacent(id int64) (prev, next *model.Article) {
	prev = &model.Article{}
	if err := s.DB.QueryRow("SELECT id,title,slug FROM articles WHERE id<? AND status='published' ORDER BY id DESC LIMIT 1", id).Scan(&prev.ID, &prev.Title, &prev.Slug); err != nil {
		prev = nil
	}
	next = &model.Article{}
	if err := s.DB.QueryRow("SELECT id,title,slug FROM articles WHERE id>? AND status='published' ORDER BY id ASC LIMIT 1", id).Scan(&next.ID, &next.Title, &next.Slug); err != nil {
		next = nil
	}
	return
}
