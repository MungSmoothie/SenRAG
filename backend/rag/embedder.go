package rag

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type Embedder struct {
	client *openai.Client
	cfg    interface{}
}

func NewEmbedder(baseURL, apiKey, model string) *Embedder {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return &Embedder{
		client: openai.NewClientWithConfig(config),
		cfg: map[string]interface{}{
			"model": model,
		},
	}
}

func (e *Embedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: e.cfg.(map[string]interface{})["model"].(string),
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return resp.Data[0].Embedding, nil
}

func (e *Embedder) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: e.cfg.(map[string]interface{})["model"].(string),
		Input: texts,
	})
	if err != nil {
		return nil, err
	}
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}
	return embeddings, nil
}

func ExtractTextFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	ext := strings.ToLower(path[strings.LastIndex(path, "."):])

	switch ext {
	case ".txt", ".md", ".json", ".yaml", ".yml", ".toml", ".go", ".py", ".java", ".js", ".ts", ".html", ".css", ".xml":
		return content, nil
	case ".pdf":
		// PDF处理需要额外库，这里简化处理
		return content[:min(10000, len(content))], nil
	default:
		// 尝试读取前10KB
		return content[:min(10000, len(content))], nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Embedder) GetEmbeddingDimensions() int {
	return 1536 // OpenAI text-embedding-3-small
}
