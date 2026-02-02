package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"senrag/config"
	"senrag/rag"
)

type RAGService struct {
	embedder *rag.Embedder
	llm      *rag.LLM
	store    *rag.VectorStore
	cfg      *config.Config
}

func NewRAGService(cfg *config.Config) (*RAGService, error) {
	svc := &RAGService{cfg: cfg}

	// 初始化 embedder
	svc.embedder = rag.NewEmbedder(
		cfg.LLM.BaseURL,
		cfg.LLM.APIKey,
		cfg.Embedder.Model,
	)

	// 初始化 LLM
	svc.llm = rag.NewLLM(
		cfg.LLM.BaseURL,
		cfg.LLM.APIKey,
		cfg.LLM.Model,
	)

	// 初始化向量存储
	var err error
	svc.store, err = rag.NewVectorStore(
		cfg.Qdrant.Host,
		cfg.Qdrant.Port,
		cfg.Qdrant.APIKey,
		cfg.Qdrant.CollectionName,
		svc.embedder.GetEmbeddingDimensions(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init vector store: %w", err)
	}

	return svc, nil
}

func (s *RAGService) IngestDocument(ctx context.Context, filePath string) error {
	// 读取文件
	content, err := rag.ExtractTextFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 分割成chunk
	chunks := s.splitIntoChunks(content)

	// 生成embeddings
	embeddings, err := s.embedder.EmbedTexts(ctx, chunks)
	if err != nil {
		return fmt.Errorf("failed to embed texts: %w", err)
	}

	// 创建文档
	docs := make([]rag.Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = rag.Document{
			Content: chunk,
			Source:  filepath.Base(filePath),
			Vector:  embeddings[i],
		}
	}

	// 存入向量库
	return s.store.AddDocuments(ctx, docs)
}

func (s *RAGService) splitIntoChunks(content string) []string {
	const (
		chunkSize    = 1000
		chunkOverlap = 200
	)

	var chunks []string
	lines := strings.Split(content, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line) > chunkSize {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				// 保留overlap部分
				content := currentChunk.String()
				if len(content) > chunkOverlap {
					currentChunk.Reset()
					currentChunk.WriteString(content[len(content)-chunkOverlap:])
				} else {
					currentChunk.Reset()
				}
			}
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	if len(chunks) == 0 {
		return []string{content}
	}
	return chunks
}

func (s *RAGService) Query(ctx context.Context, question string, topK int) (*rag.SearchResult, string, error) {
	// 查询向量库
	queryVec, err := s.embedder.EmbedText(ctx, question)
	if err != nil {
		return nil, "", fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := s.store.Search(ctx, queryVec, topK)
	if err != nil {
		return nil, "", fmt.Errorf("failed to search: %w", err)
	}

	if len(results) == 0 {
		return nil, "未找到相关文档", nil
	}

	// 构建context
	var context strings.Builder
	for i, r := range results {
		context.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, r.Source, r.Content))
	}

	// 调用LLM
	systemPrompt := `你是一个知识库助手，请根据提供的上下文回答用户的问题。
如果上下文中没有相关信息，请明确说明。
请用中文回答。`

	userMsg := fmt.Sprintf("上下文:\n%s\n\n问题: %s", context.String(), question)

	answer, err := s.llm.Chat(ctx, []rag.ChatMessage{
		{Role: "user", Content: userMsg},
	}, systemPrompt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to chat: %w", err)
	}

	return &results[0], answer, nil
}

func (s *RAGService) StreamQuery(ctx context.Context, question string, topK int) (<-chan string, error) {
	queryVec, err := s.embedder.EmbedText(ctx, question)
	if err != nil {
		return nil, err
	}

	results, err := s.store.Search(ctx, queryVec, topK)
	if err != nil {
		return nil, err
	}

	var context strings.Builder
	for i, r := range results {
		context.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, r.Source, r.Content))
	}

	systemPrompt := `你是一个知识库助手，请根据提供的上下文回答用户的问题。
如果上下文中没有相关信息，请明确说明。
请用中文回答。`

	userMsg := fmt.Sprintf("上下文:\n%s\n\n问题: %s", context.String(), question)

	return s.llm.StreamChat(ctx, []rag.ChatMessage{
		{Role: "user", Content: userMsg},
	}, systemPrompt)
}

func (s *RAGService) CreateUploadPath() error {
	return os.MkdirAll(s.cfg.Upload.SavePath, 0755)
}

func (s *RAGService) GetUploadPath() string {
	return s.cfg.Upload.SavePath
}
