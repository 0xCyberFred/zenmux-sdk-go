package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

func TestResponseServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"id":         "resp-123",
			"object":     "response",
			"model":      "openai/gpt-4.1",
			"status":     "completed",
			"created_at": 1700000000,
			"output":     []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Responses.Create(context.Background(), responses.ResponseNewParams{
		Model: shared.ResponsesModel("openai/gpt-4.1"),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Hello"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "resp-123" {
		t.Errorf("unexpected ID: got %q, want %q", result.ID, "resp-123")
	}
	if result.Status != "completed" {
		t.Errorf("unexpected status: got %q, want %q", result.Status, "completed")
	}
}

func TestResponseServiceCreateError(t *testing.T) {
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
	_, err := client.Responses.Create(context.Background(), responses.ResponseNewParams{
		Model: shared.ResponsesModel("openai/gpt-4.1"),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Hello"),
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthError(err) {
		t.Errorf("expected auth error, got: %v", err)
	}
}
