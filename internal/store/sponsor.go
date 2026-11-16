package store

import (
	"database/sql"

	"panda-site/internal/model"
)

type SponsorStore struct{ DB *sql.DB }

func (s *SponsorStore) List(visibleOnly bool) ([]model.Sponsor, error) {
	query := `SELECT id,name,amount,currency,message,url,anonymous,visible,sponsored_at,created_at FROM sponsors`
	if visibleOnly {
		query += ` WHERE visible=1`
	}
	query += ` ORDER BY sponsored_at DESC,id DESC`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Sponsor, 0)
	for rows.Next() {
		var sponsor model.Sponsor
		if err := rows.Scan(&sponsor.ID, &sponsor.Name, &sponsor.Amount, &sponsor.Currency, &sponsor.Message, &sponsor.URL, &sponsor.Anonymous, &sponsor.Visible, &sponsor.SponsoredAt, &sponsor.CreatedAt); err != nil {
			return nil, err
		}
		if visibleOnly && sponsor.Anonymous {
			sponsor.Name = "匿名道友"
			sponsor.URL = ""
		}
		list = append(list, sponsor)
	}
	return list, rows.Err()
}

func (s *SponsorStore) Create(sponsor *model.Sponsor) error {
	res, err := s.DB.Exec(`INSERT INTO sponsors (name,amount,currency,message,url,anonymous,visible,sponsored_at) VALUES (?,?,?,?,?,?,?,?)`,
		sponsor.Name, sponsor.Amount, sponsor.Currency, sponsor.Message, sponsor.URL, sponsor.Anonymous, sponsor.Visible, sponsor.SponsoredAt)
	if err != nil {
		return err
	}
	sponsor.ID, err = res.LastInsertId()
	return err
}

func (s *SponsorStore) Update(sponsor *model.Sponsor) (bool, error) {
	res, err := s.DB.Exec(`UPDATE sponsors SET name=?,amount=?,currency=?,message=?,url=?,anonymous=?,visible=?,sponsored_at=? WHERE id=?`,
		sponsor.Name, sponsor.Amount, sponsor.Currency, sponsor.Message, sponsor.URL, sponsor.Anonymous, sponsor.Visible, sponsor.SponsoredAt, sponsor.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *SponsorStore) Delete(id int64) (bool, error) {
	res, err := s.DB.Exec(`DELETE FROM sponsors WHERE id=?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
