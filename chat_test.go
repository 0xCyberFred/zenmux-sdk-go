package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestChatServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"model":   "openai/gpt-4.1",
			"choices": []map[string]any{
				{
					"index":         0,
					"finish_reason": "stop",
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello!",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 2,
				"total_tokens":      7,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hi"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "chatcmpl-123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}

func TestChatServiceCreateError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid api key",
				"type":    "authentication_error",
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("bad-key", WithBaseURL(ProviderOpenAI, server.URL))
	_, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hi"),
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthError(err) {
		t.Errorf("expected auth error, got: %v", err)
	}
}
