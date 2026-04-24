package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponseService provides access to the Responses API.
type ResponseService struct {
	client openai.Client
}

func newResponseService(client openai.Client) *ResponseService {
	return &ResponseService{client: client}
}

// Create sends a response request and returns the result.
func (s *ResponseService) Create(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error) {
	result, err := s.client.Responses.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}

// ResponseStream wraps a streaming response from the Responses API.
type ResponseStream struct {
	stream openaiStream[responses.ResponseStreamEventUnion]
}

// CreateStream initiates a streaming response request.
func (s *ResponseService) CreateStream(ctx context.Context, params responses.ResponseNewParams) *ResponseStream {
	stream := s.client.Responses.NewStreaming(ctx, params)
	return &ResponseStream{stream: stream}
}

// Next advances to the next event in the stream. Returns false when the stream
// is exhausted or an error has occurred.
func (s *ResponseStream) Next() bool {
	return s.stream.Next()
}

// Current returns the most recently decoded stream event.
func (s *ResponseStream) Current() responses.ResponseStreamEventUnion {
	return s.stream.Current()
}

// Err returns the first error encountered during streaming, wrapped as a
// zenmux Error when applicable.
func (s *ResponseStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}

// Close terminates the underlying stream.
func (s *ResponseStream) Close() error {
	if err := s.stream.Close(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}
