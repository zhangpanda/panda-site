package store

import (
	"database/sql"
	"panda-site/internal/model"
)

type ProjectStore struct{ DB *sql.DB }

func (s *ProjectStore) List() []model.Project {
	rows, err := s.DB.Query(`SELECT id,name,description,url,github,icon,tags,group_name,sort_order,visible,created_at FROM projects ORDER BY sort_order,id`)
	if err != nil {
		return []model.Project{}
	}
	defer rows.Close()
	list := make([]model.Project, 0)
	for rows.Next() {
		var p model.Project
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.URL, &p.Github, &p.Icon, &p.Tags, &p.GroupName, &p.SortOrder, &p.Visible, &p.CreatedAt)
		list = append(list, p)
	}
	return list
}

func (s *ProjectStore) Create(p *model.Project) error {
	res, err := s.DB.Exec(`INSERT INTO projects (name,description,url,github,icon,tags,group_name,sort_order,visible) VALUES (?,?,?,?,?,?,?,?,?)`,
		p.Name, p.Description, p.URL, p.Github, p.Icon, p.Tags, p.GroupName, p.SortOrder, p.Visible)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (s *ProjectStore) Update(p *model.Project) error {
	_, err := s.DB.Exec(`UPDATE projects SET name=?,description=?,url=?,github=?,icon=?,tags=?,group_name=?,sort_order=?,visible=? WHERE id=?`,
		p.Name, p.Description, p.URL, p.Github, p.Icon, p.Tags, p.GroupName, p.SortOrder, p.Visible, p.ID)
	return err
}

func (s *ProjectStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM projects WHERE id=?", id)
	return err
}
