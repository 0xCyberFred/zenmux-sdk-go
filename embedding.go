package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
)

// EmbeddingService provides access to the embeddings API.
type EmbeddingService struct {
	client openai.Client
}

func newEmbeddingService(client openai.Client) *EmbeddingService {
	return &EmbeddingService{client: client}
}

// Create sends an embedding request and returns the result.
func (s *EmbeddingService) Create(ctx context.Context, params openai.EmbeddingNewParams) (*openai.CreateEmbeddingResponse, error) {
	result, err := s.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}
