package zenmux

import (
	"context"
	"iter"

	"google.golang.org/genai"
)

// GeminiService provides access to the Google Gemini API.
type GeminiService struct {
	client *genai.Client
}

func newGeminiService(client *genai.Client) *GeminiService {
	return &GeminiService{client: client}
}

// GenerateContent sends a content generation request and returns the result.
func (s *GeminiService) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	result, err := s.client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}

// GenerateContentStream initiates a streaming content generation request.
// It returns an iterator that yields response chunks.
func (s *GeminiService) GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	return func(yield func(*genai.GenerateContentResponse, error) bool) {
		for resp, err := range s.client.Models.GenerateContentStream(ctx, model, contents, config) {
			if err != nil {
				yield(nil, wrapGoogleError(err))
				return
			}
			if !yield(resp, nil) {
				return
			}
		}
	}
}

// GenerateImages sends an image generation request and returns the result.
func (s *GeminiService) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error) {
	result, err := s.client.Models.GenerateImages(ctx, model, prompt, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}

// GenerateVideos creates a long-running video generation operation.
func (s *GeminiService) GenerateVideos(ctx context.Context, model string, prompt string, image *genai.Image, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	result, err := s.client.Models.GenerateVideos(ctx, model, prompt, image, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}
