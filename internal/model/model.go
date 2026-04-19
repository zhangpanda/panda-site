package model

import "time"

type Article struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	HTML        string     `json:"html"`
	Summary     string     `json:"summary"`
	Cover       string     `json:"cover"`
	CategoryID  int64      `json:"category_id"`
	Category    string     `json:"category,omitempty"`
	Tags        string     `json:"tags"`
	Status      string     `json:"status"`
	Views       int        `json:"views"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Github      string    `json:"github"`
	Icon        string    `json:"icon"`
	Tags        string    `json:"tags"`
	GroupName   string    `json:"group_name"`
	SortOrder   int       `json:"sort_order"`
	Visible     bool      `json:"visible"`
	CreatedAt   time.Time `json:"created_at"`
}

type Category struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}
