package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/handler"
	"github.com/FireGuo1145/MaiGoDX/internal/middleware"
)

//go:embed all:web/dist
var distFS embed.FS

func main() {
	// 初始化数据库与持久化层
	database.InitDB()
	// AimeDB is a separate TCP daemon used by SDGA after successful PowerOn.
	go handler.StartAimeDB()

	mux := http.NewServeMux()

	// 注册认证、管理与卡片绑定 API 路由
	mux.HandleFunc("/api/auth/register", handler.HandleRegister)
	mux.HandleFunc("/api/auth/login", handler.HandleLogin)
	mux.HandleFunc("/api/auth/me", handler.HandleCurrentUser)
	mux.HandleFunc("/api/auth/logout", handler.HandleLogout)
	mux.HandleFunc("/api/auth/verify", handler.HandleVerifyEmail)
	mux.HandleFunc("/api/terminal/list", handler.HandleUserTerminals)
	mux.HandleFunc("/api/terminal/create", handler.HandleCreateUserTerminal)
	mux.HandleFunc("/api/terminal/update", handler.HandleUpdateUserTerminal)
	mux.HandleFunc("/api/terminal/delete", handler.HandleDeleteUserTerminal)
	mux.HandleFunc("/api/admin/users", handler.HandleAdminUsers)
	mux.HandleFunc("/api/admin/terminals", handler.HandleAdminTerminals)
	mux.HandleFunc("/api/admin/terminal/create", handler.HandleCreateTerminal)
	mux.HandleFunc("/api/admin/terminal/update", handler.HandleUpdateTerminal)
	mux.HandleFunc("/api/admin/terminal/delete", handler.HandleDeleteTerminal)
	mux.HandleFunc("/api/admin/events", handler.HandleAdminEvents)
	mux.HandleFunc("/api/admin/event/create", handler.HandleCreateGameEvent)
	mux.HandleFunc("/api/admin/event/update", handler.HandleUpdateGameEvent)
	mux.HandleFunc("/api/admin/event/delete", handler.HandleDeleteGameEvent)
	mux.HandleFunc("/api/admin/charges", handler.HandleAdminCharges)
	mux.HandleFunc("/api/admin/charge/create", handler.HandleCreateGameCharge)
	mux.HandleFunc("/api/admin/charge/update", handler.HandleUpdateGameCharge)
	mux.HandleFunc("/api/admin/charge/delete", handler.HandleDeleteGameCharge)
	mux.HandleFunc("/api/admin/config/get", handler.HandleGetConfigs)
	mux.HandleFunc("/api/admin/config/update", handler.HandleUpdateConfig)
	mux.HandleFunc("/api/card/bind", handler.HandleBindCard)
	mux.HandleFunc("/api/card/list", handler.HandleGetUserCards)
	mux.HandleFunc("/api/stats", handler.HandleGetStats)
	mux.HandleFunc("/sys/test", handler.HandleAllNetSelfTest)
	mux.HandleFunc("/naomitest.html", handler.HandleNaomiTest)
	mux.HandleFunc("/sys/servlet/PowerOn", handler.HandleAllNetPowerOn)
	mux.HandleFunc("/sys/servlet/DownloadOrder", handler.HandleAllNetDownloadOrder)
	mux.HandleFunc("/request", handler.HandleBillingRequest)
	mux.HandleFunc("/gs/", handler.HandleTerminalMaimai)

	// 托管前端静态资源
	subFS, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/g/") || strings.HasPrefix(path, "/api/") {
			if strings.HasPrefix(path, "/g/") {
				handler.MaimaiHandler(w, r)
			}
			return
		}

		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if _, err := fs.Stat(subFS, cleanPath); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})

	// 压缩始终启用；开发模式下额外输出全量 HTTP 访问日志。
	handlerWithMiddleware := middleware.AccessLogMiddleware(middleware.CompressionMiddleware(middleware.NormalizePathMiddleware(mux)))
	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("==================================================")
	log.Println("  MaiGoDX Server (Go-based Maimai DX Server)      ")
	log.Println("  Running on :" + port + " ...                            ")
	log.Println("==================================================")

	if err := http.ListenAndServe(":"+port, handlerWithMiddleware); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
