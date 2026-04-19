package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var funcMap = template.FuncMap{
	"html":    func(s string) template.HTML { return template.HTML(s) },
	"fmtDate": func(t *time.Time) string { if t == nil { return "" }; return t.Format("2006-01-02") },
	"fmtTime": func(t time.Time) string { return t.Format("2006-01-02") },
	"split":   func(s, sep string) []string { if s == "" { return nil }; return strings.Split(s, sep) },
	"add":     func(a, b int) int { return a + b },
	"sub":     func(a, b int) int { return a - b },
	"dict": func(pairs ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(pairs); i += 2 {
			m[pairs[i].(string)] = pairs[i+1]
		}
		return m
	},
}

func (h *Handler) PageIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	articles, _ := h.Article.List(1, 5, "published", "")
	projects := h.Project.List()
	// 只显示 visible 的项目
	visible := make([]any, 0)
	for _, p := range projects {
		if p.Visible { visible = append(visible, p) }
	}
	h.render(w, "index", map[string]any{
		"Config":   h.Config.GetAll(),
		"Articles": articles,
		"Projects": visible,
	})
}

func (h *Handler) PageProjects(w http.ResponseWriter, r *http.Request) {
	projects := h.Project.List()
	// 按 group_name 分组
	groups := make([]map[string]any, 0)
	groupMap := map[string]int{}
	for _, p := range projects {
		if !p.Visible { continue }
		g := p.GroupName
		if g == "" { g = "其他" }
		if idx, ok := groupMap[g]; ok {
			groups[idx]["items"] = append(groups[idx]["items"].([]any), p)
		} else {
			groupMap[g] = len(groups)
			groups = append(groups, map[string]any{"name": g, "items": []any{p}})
		}
	}
	h.render(w, "projects", map[string]any{
		"Config": h.Config.GetAll(),
		"Groups": groups,
	})
}

func (h *Handler) PageBlog(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 { page = 1 }
	cat := r.URL.Query().Get("category")
	articles, total := h.Article.List(page, 10, "published", cat)
	categories := h.Category.List()
	totalPages := (total + 9) / 10
	h.render(w, "blog", map[string]any{
		"Config":     h.Config.GetAll(),
		"Articles":   articles,
		"Categories": categories,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
		"Category":   cat,
	})
}

func (h *Handler) PageArticle(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	a, err := h.Article.GetBySlug(slug)
	if err != nil || a.Status != "published" {
		http.NotFound(w, r)
		return
	}
	h.Article.IncrViews(a.ID)
	a.Views++
	prev, next := h.Article.GetAdjacent(a.ID)
	h.render(w, "article", map[string]any{
		"Config":  h.Config.GetAll(),
		"Article": a,
		"Prev":    prev,
		"Next":    next,
	})
}

func (h *Handler) PageAbout(w http.ResponseWriter, r *http.Request) {
	cfg := h.Config.GetAll()
	aboutHTML := ""
	if md := cfg["about_markdown"]; md != "" {
		aboutHTML = renderMarkdown(md)
	}
	h.render(w, "about", map[string]any{
		"Config":    cfg,
		"AboutHTML": aboutHTML,
	})
}

func (h *Handler) RSS(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.Article.List(1, 20, "published", "")
	cfg := h.Config.GetAll()
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
<title>%s</title><description>%s</description><link>%s</link>`,
		template.HTMLEscapeString(cfg["title"]),
		template.HTMLEscapeString(cfg["description"]),
		template.HTMLEscapeString(cfg["url"]))
	for _, a := range articles {
		pub := ""
		if a.PublishedAt != nil { pub = a.PublishedAt.Format(time.RFC1123Z) }
		fmt.Fprintf(w, `<item><title>%s</title><link>%s/blog/%s</link><description>%s</description><pubDate>%s</pubDate></item>`,
			template.HTMLEscapeString(a.Title),
			template.HTMLEscapeString(cfg["url"]),
			template.HTMLEscapeString(a.Slug),
			template.HTMLEscapeString(a.Summary), pub)
	}
	fmt.Fprint(w, `</channel></rss>`)
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"web/templates/layout.html",
		"web/templates/"+name+".html",
	)
	if err != nil {
		http.Error(w, "模板错误: "+err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
