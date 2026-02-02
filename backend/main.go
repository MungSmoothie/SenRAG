package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"senrag/config"
	"senrag/handlers"
	"senrag/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded from: %s", cfgPath)

	// 创建上传目录
	os.MkdirAll(cfg.Upload.SavePath, 0755)

	// 初始化服务
	svc, err := services.NewRAGService(cfg)
	if err != nil {
		log.Fatalf("Failed to create RAG service: %v", err)
	}

	// 创建上传目录
	if err := svc.CreateUploadPath(); err != nil {
		log.Fatalf("Failed to create upload path: %v", err)
	}

	// 启动服务
	r := gin.Default()
	h := handlers.NewHandler(svc)
	h.SetupRoutes(r)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting SenRAG server on %s", addr)

	// 优雅关闭
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
