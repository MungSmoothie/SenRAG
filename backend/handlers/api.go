package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"senrag/config"
	"senrag/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *services.RAGService
}

func NewHandler(svc *services.RAGService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.POST("/upload", h.Upload)
		api.POST("/query", h.Query)
		api.POST("/query/stream", h.StreamQuery)
		api.GET("/config", h.GetConfig)
		api.PUT("/config", h.UpdateConfig)
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查扩展名
	ext := filepath.Ext(file.Filename)
	allowed := h.svc.cfg.Upload.AllowedExts
	if len(allowed) > 0 {
		allowedExt := false
		for _, e := range allowed {
			if strings.EqualFold(ext, e) {
				allowedExt = true
				break
			}
		}
		if !allowedExt {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不支持的文件类型",
				"allowed": allowed,
			})
			return
		}
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	savePath := filepath.Join(h.svc.GetUploadPath(), filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 处理文档
	if err := h.svc.IngestDocument(c.Request.Context(), savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "上传成功",
		"filename": file.Filename,
		"id": filename,
	})
}

func (h *Handler) Query(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
		TopK     int    `json:"top_k"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}

	_, answer, err := h.svc.Query(c.Request.Context(), req.Question, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer": answer,
	})
}

func (h *Handler) StreamQuery(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
		TopK     int    `json:"top_k"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}

	stream, err := h.svc.StreamQuery(c.Request.Context(), req.Question, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for chunk := range stream {
		c.SSEvent("", chunk)
		c.Writer.Flush()
	}
	c.SSEvent("", "[DONE]")
}

func (h *Handler) GetConfig(c *gin.Context) {
	cfg := h.svc.cfg
	c.JSON(http.StatusOK, gin.H{
		"llm": gin.H{
			"provider":   cfg.LLM.Provider,
			"base_url":   cfg.LLM.BaseURL,
			"model":      cfg.LLM.Model,
			"max_tokens": cfg.LLM.MaxTokens,
		},
		"qdrant": gin.H{
			"host":     cfg.Qdrant.Host,
			"port":     cfg.Qdrant.Port,
		},
		"upload": gin.H{
			"max_file_size": cfg.Upload.MaxFileSize,
			"allowed_extensions": cfg.Upload.AllowedExts,
		},
	})
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	// 更新配置（不持久化，仅当前实例）
	var req struct {
		LLM struct {
			BaseURL string `json:"base_url"`
			Model   string `json:"model"`
		} `json:"llm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.LLM.BaseURL != "" {
		h.svc.cfg.LLM.BaseURL = req.LLM.BaseURL
	}
	if req.LLM.Model != "" {
		h.svc.cfg.LLM.Model = req.LLM.Model
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}
