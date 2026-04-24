package zenmux

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
)

// MessageService provides access to the Anthropic Messages API.
type MessageService struct {
	client anthropic.Client
}

func newMessageService(client anthropic.Client) *MessageService {
	return &MessageService{client: client}
}

// Create sends a message request and returns the result.
func (s *MessageService) Create(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	result, err := s.client.Messages.New(ctx, params)
	if err != nil {
		return nil, wrapAnthropicError(err)
	}
	return result, nil
}

// MessageStream wraps a streaming message response from the Anthropic API.
type MessageStream struct {
	stream anthropicStream
}

// CreateStream initiates a streaming message request.
func (s *MessageService) CreateStream(ctx context.Context, params anthropic.MessageNewParams) *MessageStream {
	stream := s.client.Messages.NewStreaming(ctx, params)
	return &MessageStream{stream: stream}
}

// Next advances to the next event in the stream. Returns false when the stream
// is exhausted or an error has occurred.
func (s *MessageStream) Next() bool { return s.stream.Next() }

// Current returns the most recently decoded stream event.
func (s *MessageStream) Current() anthropic.MessageStreamEventUnion { return s.stream.Current() }

// Err returns the first error encountered during streaming, wrapped as a
// zenmux Error when applicable.
func (s *MessageStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapAnthropicError(err)
	}
	return nil
}

// Close terminates the underlying stream.
func (s *MessageStream) Close() error {
	if err := s.stream.Close(); err != nil {
		return wrapAnthropicError(err)
	}
	return nil
}

// anthropicStream is an interface matching the ssestream.Stream pattern from
// anthropic-sdk-go, so we don't need to import their internal ssestream package.
type anthropicStream interface {
	Next() bool
	Current() anthropic.MessageStreamEventUnion
	Err() error
	Close() error
}
