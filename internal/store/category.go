package store

import (
	"database/sql"
	"panda-site/internal/model"
)

type CategoryStore struct{ DB *sql.DB }

func (s *CategoryStore) List() []model.Category {
	rows, err := s.DB.Query(`SELECT c.id,c.name,c.slug,(SELECT COUNT(*) FROM articles WHERE category_id=c.id AND status='published') FROM categories c ORDER BY c.id`)
	if err != nil {
		return []model.Category{}
	}
	defer rows.Close()
	list := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Count)
		list = append(list, c)
	}
	return list
}

func (s *CategoryStore) Create(c *model.Category) error {
	res, err := s.DB.Exec("INSERT INTO categories (name,slug) VALUES (?,?)", c.Name, c.Slug)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (s *CategoryStore) Update(c *model.Category) error {
	_, err := s.DB.Exec("UPDATE categories SET name=?,slug=? WHERE id=?", c.Name, c.Slug, c.ID)
	return err
}

func (s *CategoryStore) Delete(id int64) error {
	s.DB.Exec("UPDATE articles SET category_id=0 WHERE category_id=?", id)
	_, err := s.DB.Exec("DELETE FROM categories WHERE id=?", id)
	return err
}
