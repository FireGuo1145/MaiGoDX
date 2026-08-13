package main

import (
	"log"
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/handler"
	"github.com/FireGuo1145/MaiGoDX/internal/middleware"
)

func main() {
	// 初始化 SQLite 数据库与持久化层
	database.InitDB()

	mux := http.NewServeMux()

	// 注册路由，兼容 SDGA (国际版) 与 SDEZ (日版) 的 maimai2 路由
	mux.HandleFunc("/g/SDGA/", handler.MaimaiHandler)
	mux.HandleFunc("/g/SDEZ/", handler.MaimaiHandler)

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
