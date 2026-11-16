package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"panda-site/internal/handler"
	"panda-site/internal/middleware"
	"panda-site/internal/store"
)

func main() {
	addr := flag.String("addr", ":8090", "listen address")
	dbPath := flag.String("db", "site.db", "sqlite database path")
	flag.Parse()

	db, err := store.InitDB(*dbPath)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer db.Close()

	h := &handler.Handler{
		Article:    &store.ArticleStore{DB: db},
		Project:    &store.ProjectStore{DB: db},
		Category:   &store.CategoryStore{DB: db},
		Config:     &store.ConfigStore{DB: db},
		FriendLink: &store.FriendLinkStore{DB: db},
		Sponsor:    &store.SponsorStore{DB: db},
		DB:         db,
	}

	mux := http.NewServeMux()

	// 页面（服务端渲染）
	mux.HandleFunc("GET /{$}", h.PageIndex)
	mux.HandleFunc("GET /projects", h.PageProjects)
	mux.HandleFunc("GET /links", h.PageLinks)
	mux.HandleFunc("GET /sponsor", h.PageSponsor)
	mux.HandleFunc("GET /blog", h.PageBlog)
	mux.HandleFunc("GET /blog/{slug}", h.PageArticle)
	mux.HandleFunc("GET /about", h.PageAbout)
	mux.HandleFunc("GET /rss", h.RSS)

	// 公开 API
	mux.HandleFunc("GET /api/articles", h.ListArticles)
	mux.HandleFunc("GET /api/articles/{slug}", h.GetArticle)
	mux.HandleFunc("GET /api/projects", h.ListProjects)
	mux.HandleFunc("GET /api/links", h.ListFriendLinks)
	mux.HandleFunc("GET /api/sponsors", h.ListSponsors)
	mux.HandleFunc("GET /api/categories", h.ListCategories)
	mux.HandleFunc("GET /api/config", h.GetConfig)

	// 认证
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("POST /api/init", h.InitAdmin)

	// 管理 API（需要认证）
	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/articles", h.AdminListArticles)
	admin.HandleFunc("GET /api/admin/articles/{id}", h.AdminGetArticle)
	admin.HandleFunc("POST /api/admin/articles", h.CreateArticle)
	admin.HandleFunc("PUT /api/admin/articles/{id}", h.UpdateArticle)
	admin.HandleFunc("DELETE /api/admin/articles/{id}", h.DeleteArticle)
	admin.HandleFunc("POST /api/admin/projects", h.CreateProject)
	admin.HandleFunc("PUT /api/admin/projects/{id}", h.UpdateProject)
	admin.HandleFunc("DELETE /api/admin/projects/{id}", h.DeleteProject)
	admin.HandleFunc("GET /api/admin/links", h.AdminListFriendLinks)
	admin.HandleFunc("POST /api/admin/links", h.CreateFriendLink)
	admin.HandleFunc("PUT /api/admin/links/{id}", h.UpdateFriendLink)
	admin.HandleFunc("DELETE /api/admin/links/{id}", h.DeleteFriendLink)
	admin.HandleFunc("GET /api/admin/sponsors", h.AdminListSponsors)
	admin.HandleFunc("POST /api/admin/sponsors", h.CreateSponsor)
	admin.HandleFunc("PUT /api/admin/sponsors/{id}", h.UpdateSponsor)
	admin.HandleFunc("DELETE /api/admin/sponsors/{id}", h.DeleteSponsor)
	admin.HandleFunc("POST /api/admin/categories", h.CreateCategory)
	admin.HandleFunc("PUT /api/admin/categories/{id}", h.UpdateCategory)
	admin.HandleFunc("DELETE /api/admin/categories/{id}", h.DeleteCategory)
	admin.HandleFunc("PUT /api/admin/config", h.UpdateConfig)
	admin.HandleFunc("POST /api/admin/upload", h.Upload)
	mux.Handle("/api/admin/", middleware.Auth(admin))

	// 静态文件
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("web/admin"))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	srv := &http.Server{Addr: *addr, Handler: cors(mux)}
	log.Printf("🐼 Panda Site 启动 %s", *addr)
	log.Printf("   站点: http://localhost%s", *addr)
	log.Printf("   后台: http://localhost%s/admin/", *addr)

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println("正在关闭...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
