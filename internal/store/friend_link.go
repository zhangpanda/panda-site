package store

import (
	"database/sql"

	"panda-site/internal/model"
)

type FriendLinkStore struct{ DB *sql.DB }

func (s *FriendLinkStore) List(visibleOnly bool) ([]model.FriendLink, error) {
	query := `SELECT id,name,url,logo,description,group_name,sort_order,visible,created_at FROM friend_links`
	if visibleOnly {
		query += ` WHERE visible=1`
	}
	query += ` ORDER BY sort_order,id`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.FriendLink, 0)
	for rows.Next() {
		var link model.FriendLink
		if err := rows.Scan(&link.ID, &link.Name, &link.URL, &link.Logo, &link.Description, &link.GroupName, &link.SortOrder, &link.Visible, &link.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, link)
	}
	return list, rows.Err()
}

func (s *FriendLinkStore) Create(link *model.FriendLink) error {
	res, err := s.DB.Exec(`INSERT INTO friend_links (name,url,logo,description,group_name,sort_order,visible) VALUES (?,?,?,?,?,?,?)`,
		link.Name, link.URL, link.Logo, link.Description, link.GroupName, link.SortOrder, link.Visible)
	if err != nil {
		return err
	}
	link.ID, err = res.LastInsertId()
	return err
}

func (s *FriendLinkStore) Update(link *model.FriendLink) (bool, error) {
	res, err := s.DB.Exec(`UPDATE friend_links SET name=?,url=?,logo=?,description=?,group_name=?,sort_order=?,visible=? WHERE id=?`,
		link.Name, link.URL, link.Logo, link.Description, link.GroupName, link.SortOrder, link.Visible, link.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *FriendLinkStore) Delete(id int64) (bool, error) {
	res, err := s.DB.Exec(`DELETE FROM friend_links WHERE id=?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
