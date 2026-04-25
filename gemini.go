package zenmux

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"

	"google.golang.org/genai"
)

// GeminiService provides access to the Google Gemini API.
type GeminiService struct {
	client *genai.Client
	cfg    *config
}

func newGeminiService(client *genai.Client, cfg *config) *GeminiService {
	return &GeminiService{client: client, cfg: cfg}
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

// GetVideosOperation polls a long-running video generation operation.
//
// The genai library's Operations.GetVideosOperation does not work with ZenMux
// because ZenMux's /api/vertex-ai endpoint speaks the Vertex AI protocol
// (POST <resource>:fetchPredictOperation) but expects API version v1beta
// rather than the Vertex default v1beta1. This method issues the correct
// request directly.
func (s *GeminiService) GetVideosOperation(ctx context.Context, op *genai.GenerateVideosOperation) (*genai.GenerateVideosOperation, error) {
	if op == nil || op.Name == "" {
		return nil, fmt.Errorf("operation name is empty")
	}
	parts := strings.SplitN(op.Name, "/operations/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid operation name: %q", op.Name)
	}
	resourceName := parts[0]

	url := fmt.Sprintf("%s/v1beta/%s:fetchPredictOperation", s.cfg.baseURL(ProviderGoogle), resourceName)
	body, err := json.Marshal(map[string]string{"operationName": op.Name})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.apiKey)
	req.Header.Set("Content-Type", "application/json")

	hc := s.cfg.httpClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{
			Provider:   ProviderGoogle,
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return decodeVideoOperation(respBody)
}

// vertexVideoOperation mirrors the Vertex AI :fetchPredictOperation response
// shape. The `videos[]` payload uses `bytesBase64Encoded`, which we decode
// into genai.Video.VideoBytes.
type vertexVideoOperation struct {
	Name     string         `json:"name"`
	Done     bool           `json:"done"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Error    map[string]any `json:"error,omitempty"`
	Response *struct {
		RAIMediaFilteredCount   int32    `json:"raiMediaFilteredCount,omitempty"`
		RAIMediaFilteredReasons []string `json:"raiMediaFilteredReasons,omitempty"`
		Videos                  []struct {
			GCSURI             string `json:"gcsUri,omitempty"`
			BytesBase64Encoded string `json:"bytesBase64Encoded,omitempty"`
			MIMEType           string `json:"mimeType,omitempty"`
		} `json:"videos,omitempty"`
	} `json:"response,omitempty"`
}

func decodeVideoOperation(body []byte) (*genai.GenerateVideosOperation, error) {
	var raw vertexVideoOperation
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing operation response: %w", err)
	}

	out := &genai.GenerateVideosOperation{
		Name:     raw.Name,
		Done:     raw.Done,
		Metadata: raw.Metadata,
		Error:    raw.Error,
	}
	if raw.Response == nil {
		return out, nil
	}

	videos := make([]*genai.GeneratedVideo, 0, len(raw.Response.Videos))
	for _, v := range raw.Response.Videos {
		gv := &genai.Video{URI: v.GCSURI, MIMEType: v.MIMEType}
		if v.BytesBase64Encoded != "" {
			decoded, err := base64.StdEncoding.DecodeString(v.BytesBase64Encoded)
			if err != nil {
				return nil, fmt.Errorf("decoding video bytes: %w", err)
			}
			gv.VideoBytes = decoded
		}
		videos = append(videos, &genai.GeneratedVideo{Video: gv})
	}
	out.Response = &genai.GenerateVideosResponse{
		GeneratedVideos:         videos,
		RAIMediaFilteredCount:   raw.Response.RAIMediaFilteredCount,
		RAIMediaFilteredReasons: raw.Response.RAIMediaFilteredReasons,
	}
	return out, nil
}
