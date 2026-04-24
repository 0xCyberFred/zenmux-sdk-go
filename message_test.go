package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestMessageServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("unexpected x-api-key header: %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("expected anthropic-version header to be present")
		}

		resp := map[string]any{
			"id":    "msg_123",
			"type":  "message",
			"role":  "assistant",
			"model": "claude-sonnet-4-20250514",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Hello!",
				},
			},
			"stop_reason": "end_turn",
			"usage": map[string]any{
				"input_tokens":  10,
				"output_tokens": 3,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderAnthropic, server.URL))
	result, err := client.Messages.Create(context.Background(), anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "msg_123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}

func TestMessageServiceCreateError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "rate_limit_error",
				"message": "rate limit exceeded",
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderAnthropic, server.URL))
	_, err := client.Messages.Create(context.Background(), anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsRateLimitError(err) {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}
