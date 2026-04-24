package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
)

// ChatService provides access to the chat completions API.
type ChatService struct {
	client openai.Client
}

func newChatService(client openai.Client) *ChatService {
	return &ChatService{client: client}
}

// Create sends a chat completion request and returns the result.
func (s *ChatService) Create(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	result, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}

// ChatCompletionStream wraps a streaming chat completion response.
type ChatCompletionStream struct {
	stream openaiStream[openai.ChatCompletionChunk]
}

// CreateStream initiates a streaming chat completion request.
func (s *ChatService) CreateStream(ctx context.Context, params openai.ChatCompletionNewParams) *ChatCompletionStream {
	stream := s.client.Chat.Completions.NewStreaming(ctx, params)
	return &ChatCompletionStream{stream: stream}
}

// Next advances to the next chunk in the stream. Returns false when the stream
// is exhausted or an error has occurred.
func (s *ChatCompletionStream) Next() bool {
	return s.stream.Next()
}

// Current returns the most recently decoded chunk.
func (s *ChatCompletionStream) Current() openai.ChatCompletionChunk {
	return s.stream.Current()
}

// Err returns the first error encountered during streaming, wrapped as a
// zenmux Error when applicable.
func (s *ChatCompletionStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}

// Close terminates the underlying stream.
func (s *ChatCompletionStream) Close() error {
	return s.stream.Close()
}

// openaiStream is a generic interface matching the ssestream.Stream pattern
// from openai-go, so we don't need to import their internal ssestream package.
type openaiStream[T any] interface {
	Next() bool
	Current() T
	Err() error
	Close() error
}
