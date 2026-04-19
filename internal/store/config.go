package store

import "database/sql"

type ConfigStore struct{ DB *sql.DB }

func (s *ConfigStore) Get(key string) string {
	var v string
	s.DB.QueryRow("SELECT value FROM site_config WHERE key=?", key).Scan(&v)
	return v
}

func (s *ConfigStore) GetAll() map[string]string {
	rows, err := s.DB.Query("SELECT key,value FROM site_config")
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		m[k] = v
	}
	return m
}

func (s *ConfigStore) Set(key, value string) {
	s.DB.Exec("INSERT INTO site_config (key,value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=?", key, value, value)
}

func (s *ConfigStore) SetMulti(m map[string]string) {
	for k, v := range m {
		s.Set(k, v)
	}
}
