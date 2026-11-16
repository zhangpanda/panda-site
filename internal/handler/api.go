package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"panda-site/internal/middleware"
	"panda-site/internal/model"
	"panda-site/internal/store"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

type Handler struct {
	Article    *store.ArticleStore
	Project    *store.ProjectStore
	Category   *store.CategoryStore
	Config     *store.ConfigStore
	FriendLink *store.FriendLinkStore
	Sponsor    *store.SponsorStore
	DB         *sql.DB
}

func jsonOK(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, code int, msg string) {
	jsonOK(w, code, map[string]string{"error": msg})
}

func pageParams(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	return page, size
}

// === 公开 API ===

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	list, total := h.Article.List(page, size, "published", r.URL.Query().Get("category"))
	jsonOK(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	a, err := h.Article.GetBySlug(r.PathValue("slug"))
	if err != nil || a.Status != "published" {
		jsonErr(w, 404, "文章不存在")
		return
	}
	h.Article.IncrViews(a.ID)
	a.Views++
	jsonOK(w, 200, a)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, 200, h.Project.List())
}

func (h *Handler) ListFriendLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.FriendLink.List(true)
	if err != nil {
		jsonErr(w, 500, "读取友情链接失败")
		return
	}
	jsonOK(w, 200, links)
}

func (h *Handler) ListSponsors(w http.ResponseWriter, r *http.Request) {
	sponsors, err := h.Sponsor.List(true)
	if err != nil {
		jsonErr(w, 500, "读取赞助记录失败")
		return
	}
	jsonOK(w, 200, sponsors)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, 200, h.Category.List())
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, 200, h.Config.GetAll())
}

// === 认证 ===

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	var hash string
	h.DB.QueryRow("SELECT password_hash FROM admin WHERE username=?", req.Username).Scan(&hash)
	if hash == "" || !middleware.CheckPassword(hash, req.Password) {
		jsonErr(w, 401, "用户名或密码错误")
		return
	}
	token, _ := middleware.GenerateToken(req.Username)
	jsonOK(w, 200, map[string]string{"token": token})
}

func (h *Handler) InitAdmin(w http.ResponseWriter, r *http.Request) {
	var count int
	h.DB.QueryRow("SELECT COUNT(*) FROM admin").Scan(&count)
	if count > 0 {
		jsonErr(w, 400, "管理员已存在")
		return
	}
	var req struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if req.Username == "" || len(req.Password) < 6 {
		jsonErr(w, 400, "用户名不能为空，密码至少6位")
		return
	}
	h.DB.Exec("INSERT INTO admin (username,password_hash) VALUES (?,?)", req.Username, middleware.HashPassword(req.Password))
	token, _ := middleware.GenerateToken(req.Username)
	jsonOK(w, 201, map[string]string{"token": token})
}

// === 管理 API ===

func (h *Handler) AdminListArticles(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	list, total := h.Article.List(page, size, r.URL.Query().Get("status"), "")
	jsonOK(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) AdminGetArticle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	a, err := h.Article.GetByID(id)
	if err != nil {
		jsonErr(w, 404, "文章不存在")
		return
	}
	jsonOK(w, 200, a)
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var a model.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if a.Title == "" {
		jsonErr(w, 400, "标题不能为空")
		return
	}
	if a.Slug == "" {
		a.Slug = slugify(a.Title)
	}
	a.HTML = renderMarkdown(a.Content)
	if a.Summary == "" {
		a.Summary = extractSummary(a.HTML, 200)
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	if a.Status == "published" {
		now := time.Now()
		a.PublishedAt = &now
	}
	if err := h.Article.Create(&a); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonOK(w, 201, a)
}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	old, err := h.Article.GetByID(id)
	if err != nil {
		jsonErr(w, 404, "文章不存在")
		return
	}
	var req struct {
		Title      *string `json:"title"`
		Slug       *string `json:"slug"`
		Content    *string `json:"content"`
		Summary    *string `json:"summary"`
		Cover      *string `json:"cover"`
		CategoryID *int64  `json:"category_id"`
		Tags       *string `json:"tags"`
		Status     *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	a := *old
	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Slug != nil {
		a.Slug = *req.Slug
	}
	if req.Content != nil {
		a.Content = *req.Content
		a.HTML = renderMarkdown(a.Content)
	}
	if req.Summary != nil {
		a.Summary = *req.Summary
	}
	if req.Cover != nil {
		a.Cover = *req.Cover
	}
	if req.CategoryID != nil {
		a.CategoryID = *req.CategoryID
	}
	if req.Tags != nil {
		a.Tags = *req.Tags
	}
	if req.Status != nil {
		a.Status = *req.Status
	}
	if req.Content != nil && (req.Summary == nil || *req.Summary == "") {
		a.Summary = extractSummary(a.HTML, 200)
	}
	if a.Status == "published" && old.Status != "published" {
		now := time.Now()
		a.PublishedAt = &now
	}
	h.Article.Update(&a)
	jsonOK(w, 200, a)
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Article.Delete(id)
	jsonOK(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var p model.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if p.Name == "" {
		jsonErr(w, 400, "项目名称不能为空")
		return
	}
	h.Project.Create(&p)
	jsonOK(w, 201, p)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var p model.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	p.ID = id
	h.Project.Update(&p)
	jsonOK(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Project.Delete(id)
	jsonOK(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) AdminListFriendLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.FriendLink.List(false)
	if err != nil {
		jsonErr(w, 500, "读取友情链接失败")
		return
	}
	jsonOK(w, 200, links)
}

func (h *Handler) CreateFriendLink(w http.ResponseWriter, r *http.Request) {
	var link model.FriendLink
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if msg := validateFriendLink(&link); msg != "" {
		jsonErr(w, 400, msg)
		return
	}
	if err := h.FriendLink.Create(&link); err != nil {
		jsonErr(w, 500, "保存友情链接失败")
		return
	}
	jsonOK(w, 201, link)
}

func (h *Handler) UpdateFriendLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		jsonErr(w, 400, "无效的友情链接 ID")
		return
	}
	var link model.FriendLink
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	link.ID = id
	if msg := validateFriendLink(&link); msg != "" {
		jsonErr(w, 400, msg)
		return
	}
	updated, err := h.FriendLink.Update(&link)
	if err != nil {
		jsonErr(w, 500, "保存友情链接失败")
		return
	}
	if !updated {
		jsonErr(w, 404, "友情链接不存在")
		return
	}
	jsonOK(w, 200, link)
}

func (h *Handler) DeleteFriendLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		jsonErr(w, 400, "无效的友情链接 ID")
		return
	}
	deleted, err := h.FriendLink.Delete(id)
	if err != nil {
		jsonErr(w, 500, "删除友情链接失败")
		return
	}
	if !deleted {
		jsonErr(w, 404, "友情链接不存在")
		return
	}
	jsonOK(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) AdminListSponsors(w http.ResponseWriter, r *http.Request) {
	sponsors, err := h.Sponsor.List(false)
	if err != nil {
		jsonErr(w, 500, "读取赞助记录失败")
		return
	}
	jsonOK(w, 200, sponsors)
}

func (h *Handler) CreateSponsor(w http.ResponseWriter, r *http.Request) {
	var sponsor model.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&sponsor); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if msg := validateSponsor(&sponsor); msg != "" {
		jsonErr(w, 400, msg)
		return
	}
	if err := h.Sponsor.Create(&sponsor); err != nil {
		jsonErr(w, 500, "保存赞助记录失败")
		return
	}
	jsonOK(w, 201, sponsor)
}

func (h *Handler) UpdateSponsor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		jsonErr(w, 400, "无效的赞助记录 ID")
		return
	}
	var sponsor model.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&sponsor); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	sponsor.ID = id
	if msg := validateSponsor(&sponsor); msg != "" {
		jsonErr(w, 400, msg)
		return
	}
	updated, err := h.Sponsor.Update(&sponsor)
	if err != nil {
		jsonErr(w, 500, "保存赞助记录失败")
		return
	}
	if !updated {
		jsonErr(w, 404, "赞助记录不存在")
		return
	}
	jsonOK(w, 200, sponsor)
}

func (h *Handler) DeleteSponsor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		jsonErr(w, 400, "无效的赞助记录 ID")
		return
	}
	deleted, err := h.Sponsor.Delete(id)
	if err != nil {
		jsonErr(w, 500, "删除赞助记录失败")
		return
	}
	if !deleted {
		jsonErr(w, 404, "赞助记录不存在")
		return
	}
	jsonOK(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var c model.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	if c.Name == "" {
		jsonErr(w, 400, "名称不能为空")
		return
	}
	if c.Slug == "" {
		c.Slug = slugify(c.Name)
	}
	if err := h.Category.Create(&c); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonOK(w, 201, c)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var c model.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	c.ID = id
	h.Category.Update(&c)
	jsonOK(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Category.Delete(id)
	jsonOK(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg map[string]string
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonErr(w, 400, "请求格式错误")
		return
	}
	h.Config.SetMulti(cfg)
	jsonOK(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonErr(w, 400, "请选择文件")
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		jsonErr(w, 400, "只支持 jpg/png/gif/webp 格式")
		return
	}
	os.MkdirAll("uploads", 0755)
	name := fmt.Sprintf("%d%s", time.Now().UnixMilli(), ext)
	dst, err := os.Create(filepath.Join("uploads", name))
	if err != nil {
		jsonErr(w, 500, "保存失败")
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
	jsonOK(w, 200, map[string]string{"url": "/uploads/" + name})
}

// === 工具函数 ===

var sponsorAmountRE = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)
var currencyRE = regexp.MustCompile(`^[A-Z]{3,8}$`)

func validateFriendLink(link *model.FriendLink) string {
	link.Name = strings.TrimSpace(link.Name)
	link.URL = strings.TrimSpace(link.URL)
	link.Logo = strings.TrimSpace(link.Logo)
	link.GroupName = strings.TrimSpace(link.GroupName)
	if link.Name == "" {
		return "名称不能为空"
	}
	if !validWebURL(link.URL, false) {
		return "网址必须是有效的 http 或 https 地址"
	}
	if link.Logo != "" && !validWebURL(link.Logo, true) {
		return "Logo 必须是有效的网址或站内路径"
	}
	return ""
}

func validateSponsor(sponsor *model.Sponsor) string {
	sponsor.Name = strings.TrimSpace(sponsor.Name)
	sponsor.Amount = strings.TrimSpace(sponsor.Amount)
	sponsor.Currency = strings.ToUpper(strings.TrimSpace(sponsor.Currency))
	sponsor.URL = strings.TrimSpace(sponsor.URL)
	if sponsor.Anonymous && sponsor.Name == "" {
		sponsor.Name = "匿名道友"
	}
	if sponsor.Name == "" {
		return "赞助者名称不能为空"
	}
	if !sponsorAmountRE.MatchString(sponsor.Amount) {
		return "金额必须是最多两位小数的非负数字"
	}
	if sponsor.Currency == "" {
		sponsor.Currency = "CNY"
	}
	if !currencyRE.MatchString(sponsor.Currency) {
		return "币种应为 3 到 8 位大写字母"
	}
	if sponsor.URL != "" && !validWebURL(sponsor.URL, false) {
		return "赞助者链接必须是有效的 http 或 https 地址"
	}
	if sponsor.SponsoredAt == "" {
		sponsor.SponsoredAt = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", sponsor.SponsoredAt); err != nil {
		return "赞助日期格式应为 YYYY-MM-DD"
	}
	return ""
}

func validWebURL(raw string, allowRelative bool) bool {
	if allowRelative && strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return true
	}
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, highlighting.NewHighlighting(highlighting.WithStyle("monokai"))),
	goldmark.WithRendererOptions(goldhtml.WithUnsafe()),
)

func renderMarkdown(src string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return html.EscapeString(src)
	}
	return buf.String()
}

var reNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "post-" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	return s
}

func extractSummary(htmlStr string, maxLen int) string {
	text := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlStr, "")
	text = html.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")
	r := []rune(text)
	if len(r) > maxLen {
		return string(r[:maxLen]) + "..."
	}
	return text
}
