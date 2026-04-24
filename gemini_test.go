package zenmux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/genai"

	googleprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/google"
)

func TestZenMuxTransport(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the transport sets Bearer auth
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected Authorization 'Bearer test-key', got %q", auth)
		}
		// Verify x-goog-api-key is removed
		if got := r.Header.Get("x-goog-api-key"); got != "" {
			t.Errorf("expected x-goog-api-key to be removed, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a client through the google provider factory which applies the transport
	client, err := googleprovider.NewClient(context.Background(), "test-key", server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	// Make a request through the client — the genai SDK will build its own
	// URL path on top of the base URL. We don't care about the response
	// content; we only care that the transport rewrote the headers correctly.
	// The handler above asserts the headers.
	_, _ = client.Models.GenerateContent(context.Background(), "gemini-2.0-flash", []*genai.Content{
		{Parts: []*genai.Part{{Text: "hello"}}},
	}, nil)
}

func TestGeminiServiceGenerateContent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header from ZenMux transport
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Hello from Gemini!"},
						},
						"role": "model",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderGoogle, server.URL))
	if client.Gemini == nil {
		t.Fatal("Gemini service should not be nil")
	}

	result, err := client.Gemini.GenerateContent(
		context.Background(),
		"gemini-2.0-flash",
		[]*genai.Content{
			{Parts: []*genai.Part{{Text: "Hi"}}},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}
	if len(result.Candidates[0].Content.Parts) == 0 {
		t.Fatal("expected at least one part")
	}
	if result.Candidates[0].Content.Parts[0].Text != "Hello from Gemini!" {
		t.Errorf("unexpected text: %s", result.Candidates[0].Content.Parts[0].Text)
	}
}

func TestGeminiServiceGenerateContentError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    429,
				"message": "rate limit exceeded",
				"status":  "RESOURCE_EXHAUSTED",
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderGoogle, server.URL))
	_, err := client.Gemini.GenerateContent(
		context.Background(),
		"gemini-2.0-flash",
		[]*genai.Content{
			{Parts: []*genai.Part{{Text: "Hi"}}},
		},
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}

	var zmErr *Error
	if !errors.As(err, &zmErr) {
		t.Fatalf("expected zenmux Error, got: %T", err)
	}
	if zmErr.Provider != ProviderGoogle {
		t.Errorf("expected provider google, got %s", zmErr.Provider)
	}
}

func TestGoogleEscapeHatch(t *testing.T) {
	client := NewClient("test-key")
	gc := client.Google()
	if gc == nil {
		t.Error("Google() should return a non-nil genai.Client")
	}
}

func TestWrapGoogleErrorNil(t *testing.T) {
	if err := wrapGoogleError(nil); err != nil {
		t.Errorf("wrapGoogleError(nil) should return nil, got %v", err)
	}
}

func TestWrapGoogleErrorExtractsStatusCode(t *testing.T) {
	apiErr := &genai.APIError{
		Code:    401,
		Message: "unauthorized",
		Status:  "UNAUTHENTICATED",
	}
	wrapped := wrapGoogleError(apiErr)
	var zmErr *Error
	if !errors.As(wrapped, &zmErr) {
		t.Fatalf("expected zenmux Error, got: %T", wrapped)
	}
	if zmErr.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", zmErr.StatusCode)
	}
	if zmErr.Provider != ProviderGoogle {
		t.Errorf("expected provider google, got %s", zmErr.Provider)
	}
	if !IsAuthError(wrapped) {
		t.Error("expected IsAuthError to return true for 401")
	}
}
