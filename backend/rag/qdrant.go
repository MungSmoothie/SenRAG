package rag

import (
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/v2"
	"github.com/qdrant/go-client/v2/grpc/point"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VectorStore struct {
	client    *go_client.Client
	collection string
	dim       int
}

func NewVectorStore(host string, port int, apiKey string, collection string, dim int) (*VectorStore, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	client := go_client.NewClientConn(conn, &go_client.Config{
		APIKey: apiKey,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建collection
	err = client.CreateCollection(ctx, &go_client.CreateCollection{
		CollectionName: collection,
		VectorsConfig: &go_client.VectorsConfig{
			Config: &go_client.VectorsConfig_Params{
				Params: &go_client.DistanceVectorParams{
					Size:     uint64(dim),
					Distance: go_client.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		// 可能是已存在，忽略错误
	}

	return &VectorStore{
		client:    client,
		collection: collection,
		dim:       dim,
	}, nil
}

func (v *VectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	points := make([]*point.PointStruct, 0, len(docs))
	for i, doc := range docs {
		points = append(points, &point.PointStruct{
			Id: &point.PointId{
				Num: uint64(i + 1),
			},
			Vectors: doc.Vector,
			Payload: map[string]interface{}{
				"content": doc.Content,
				"source":  doc.Source,
				"metadata": doc.Metadata,
			},
		})
	}

	_, err := v.client.Upsert(ctx, &go_client.UpsertPoints{
		CollectionName: v.collection,
		Points:         points,
	})
	return err
}

func (v *VectorStore) Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error) {
	resp, err := v.client.Search(ctx, &go_client.SearchPoints{
		CollectionName: v.collection,
		QueryVector:    query,
		Limit:          uint64(limit),
		WithPayload:    &go_client.WithPayloadSelector{
			Selector: &go_client.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(resp.GetResult()))
	for _, r := range resp.GetResult() {
		var content, source string
		var metadata map[string]interface{}
		if p := r.GetPayload(); p != nil {
			if c, ok := p["content"].GetInterface().(string); ok {
				content = c
			}
			if s, ok := p["source"].GetInterface().(string); ok {
				source = s
			}
			if m, ok := p["metadata"].GetInterface().(string); ok {
				metadata = map[string]interface{}{"raw": m}
			}
		}
		results = append(results, SearchResult{
			Content:  content,
			Source:   source,
			Score:    float64(r.GetScore()),
			Metadata: metadata,
		})
	}
	return results, nil
}

func (v *VectorStore) DeleteCollection(ctx context.Context) error {
	_, err := v.client.DeleteCollection(ctx, v.collection)
	return err
}

type Document struct {
	Content  string
	Source   string
	Vector   []float32
	Metadata map[string]interface{}
}

type SearchResult struct {
	Content  string
	Source   string
	Score    float64
	Metadata map[string]interface{}
}
