package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/handler"
	"github.com/FireGuo1145/MaiGoDX/internal/middleware"
)

//go:embed all:web/dist
var distFS embed.FS

func main() {
	// 初始化 SQLite 数据库与持久化层
	database.InitDB()

	mux := http.NewServeMux()

	// 注册认证 API 路由
	mux.HandleFunc("/api/auth/register", handler.HandleRegister)
	mux.HandleFunc("/api/auth/login", handler.HandleLogin)
	mux.HandleFunc("/api/auth/verify", handler.HandleVerifyEmail)

	// 托管前端静态资源
	subFS, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// 如果是 maimai API 请求或认证 API 请求
		if strings.HasPrefix(path, "/g/") || strings.HasPrefix(path, "/api/") {
			if strings.HasPrefix(path, "/g/") {
				handler.MaimaiHandler(w, r)
			}
			return
		}

		// 检查静态资源是否存在，若不存在则 fallback 到 index.html (SPA 支持)
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if _, err := fs.Stat(subFS, cleanPath); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})

	// 使用压缩中间件
	handlerWithMiddleware := middleware.CompressionMiddleware(mux)

	log.Println("==================================================")
	log.Println("  MaiGoDX Server (Go-based Maimai DX Server)      ")
	log.Println("  Running on :8080 ...                            ")
	log.Println("==================================================")

	if err := http.ListenAndServe(":8080", handlerWithMiddleware); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
