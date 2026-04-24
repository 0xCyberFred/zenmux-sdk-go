package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelServiceListOpenAI(t *testing.T) {
	gte := 0.0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		resp := map[string]any{
			"object": "list",
			"data": []map[string]any{
				{
					"id":                "openai/gpt-4.1",
					"display_name":      "GPT-4.1",
					"context_length":    128000,
					"input_modalities":  []string{"text", "image"},
					"output_modalities": []string{"text"},
					"capabilities":      map[string]any{"reasoning": true},
					"pricings": map[string]any{
						"input": []map[string]any{
							{
								"value":    2.0,
								"unit":     "per_million_tokens",
								"currency": "USD",
								"conditions": map[string]any{
									"prompt_tokens": map[string]any{"gte": gte},
								},
							},
						},
					},
				},
				{
					"id":                "openai/gpt-4.1-mini",
					"display_name":      "GPT-4.1 Mini",
					"context_length":    1048576,
					"input_modalities":  []string{"text"},
					"output_modalities": []string{"text"},
					"capabilities":      map[string]any{"reasoning": false},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("test-key")
	cfg.baseURLs[ProviderOpenAI] = server.URL
	svc := newModelService(cfg)

	result, err := svc.List(context.Background(), ProviderOpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(result.Models))
	}

	m := result.Models[0]
	if m.ID != "openai/gpt-4.1" {
		t.Errorf("unexpected ID: %s", m.ID)
	}
	if m.DisplayName != "GPT-4.1" {
		t.Errorf("unexpected DisplayName: %s", m.DisplayName)
	}
	if m.Provider != ProviderOpenAI {
		t.Errorf("unexpected Provider: %s", m.Provider)
	}
	if m.ContextLength != 128000 {
		t.Errorf("unexpected ContextLength: %d", m.ContextLength)
	}
	if !m.Reasoning {
		t.Error("expected Reasoning to be true")
	}
	if len(m.InputModalities) != 2 || m.InputModalities[0] != "text" {
		t.Errorf("unexpected InputModalities: %v", m.InputModalities)
	}
	if len(m.OutputModalities) != 1 || m.OutputModalities[0] != "text" {
		t.Errorf("unexpected OutputModalities: %v", m.OutputModalities)
	}

	// Check pricing was parsed
	if m.Pricings == nil {
		t.Fatal("expected pricings to be non-nil")
	}
	inputPricing, ok := m.Pricings["input"]
	if !ok || len(inputPricing) != 1 {
		t.Fatalf("expected 1 input pricing entry, got %v", m.Pricings)
	}
	if inputPricing[0].Value != 2.0 {
		t.Errorf("unexpected pricing value: %f", inputPricing[0].Value)
	}
	if inputPricing[0].Unit != "per_million_tokens" {
		t.Errorf("unexpected pricing unit: %s", inputPricing[0].Unit)
	}
	if inputPricing[0].Conditions == nil || inputPricing[0].Conditions.PromptTokens == nil {
		t.Fatal("expected pricing conditions with prompt_tokens")
	}
	if *inputPricing[0].Conditions.PromptTokens.Gte != 0.0 {
		t.Errorf("unexpected pricing conditions gte: %f", *inputPricing[0].Conditions.PromptTokens.Gte)
	}

	// Second model should have no reasoning and no pricings
	m2 := result.Models[1]
	if m2.ID != "openai/gpt-4.1-mini" {
		t.Errorf("unexpected ID: %s", m2.ID)
	}
	if m2.Reasoning {
		t.Error("expected Reasoning to be false for mini model")
	}
	if m2.Pricings != nil {
		t.Errorf("expected nil pricings, got %v", m2.Pricings)
	}
}

func TestModelServiceListAnthropic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"data": []map[string]any{
				{
					"id":                "anthropic/claude-sonnet-4-5",
					"display_name":      "Claude Sonnet 4.5",
					"context_length":    200000,
					"input_modalities":  []string{"text", "image"},
					"output_modalities": []string{"text"},
					"capabilities":      map[string]any{"reasoning": true},
					"pricings": map[string]any{
						"input": []map[string]any{
							{
								"value":    3.0,
								"unit":     "per_million_tokens",
								"currency": "USD",
							},
						},
						"output": []map[string]any{
							{
								"value":    15.0,
								"unit":     "per_million_tokens",
								"currency": "USD",
							},
						},
					},
				},
			},
			"has_more": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("test-key")
	cfg.baseURLs[ProviderAnthropic] = server.URL
	svc := newModelService(cfg)

	result, err := svc.List(context.Background(), ProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Models))
	}

	m := result.Models[0]
	if m.ID != "anthropic/claude-sonnet-4-5" {
		t.Errorf("unexpected ID: %s", m.ID)
	}
	if m.DisplayName != "Claude Sonnet 4.5" {
		t.Errorf("unexpected DisplayName: %s", m.DisplayName)
	}
	if m.Provider != ProviderAnthropic {
		t.Errorf("unexpected Provider: %s", m.Provider)
	}
	if m.ContextLength != 200000 {
		t.Errorf("unexpected ContextLength: %d", m.ContextLength)
	}
	if !m.Reasoning {
		t.Error("expected Reasoning to be true")
	}
	if m.Pricings == nil {
		t.Fatal("expected pricings to be non-nil")
	}
	if _, ok := m.Pricings["output"]; !ok {
		t.Error("expected output pricing entry")
	}
}

func TestModelServiceListGoogle(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Google uses different field names: "name" instead of "id",
		// "displayName" instead of "display_name",
		// "inputTokenLimit" instead of "context_length".
		resp := map[string]any{
			"models": []map[string]any{
				{
					"name":              "google/gemini-2.5-pro",
					"displayName":       "Gemini 2.5 Pro",
					"inputTokenLimit":   1048576,
					"input_modalities":  []string{"text", "image", "audio", "video"},
					"output_modalities": []string{"text", "image"},
					"capabilities":      map[string]any{"reasoning": true},
					"pricings": map[string]any{
						"input": []map[string]any{
							{
								"value":    1.25,
								"unit":     "per_million_tokens",
								"currency": "USD",
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("test-key")
	cfg.baseURLs[ProviderGoogle] = server.URL
	svc := newModelService(cfg)

	result, err := svc.List(context.Background(), ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Models))
	}

	m := result.Models[0]
	// Google uses "name" field, which we map to ID
	if m.ID != "google/gemini-2.5-pro" {
		t.Errorf("unexpected ID: %s", m.ID)
	}
	// Google uses "displayName" (camelCase), which we map to DisplayName
	if m.DisplayName != "Gemini 2.5 Pro" {
		t.Errorf("unexpected DisplayName: %s", m.DisplayName)
	}
	if m.Provider != ProviderGoogle {
		t.Errorf("unexpected Provider: %s", m.Provider)
	}
	// Google uses "inputTokenLimit", which we map to ContextLength
	if m.ContextLength != 1048576 {
		t.Errorf("unexpected ContextLength: %d", m.ContextLength)
	}
	if !m.Reasoning {
		t.Error("expected Reasoning to be true")
	}
	if len(m.InputModalities) != 4 {
		t.Errorf("expected 4 input modalities, got %d", len(m.InputModalities))
	}
	if len(m.OutputModalities) != 2 {
		t.Errorf("expected 2 output modalities, got %d", len(m.OutputModalities))
	}
}

func TestModelServiceListInvalidProvider(t *testing.T) {
	cfg := defaultConfig("test-key")
	svc := newModelService(cfg)

	_, err := svc.List(context.Background(), Provider("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
	expected := "unsupported provider: invalid"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestModelServiceListHTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("bad-key")
	cfg.baseURLs[ProviderOpenAI] = server.URL
	svc := newModelService(cfg)

	_, err := svc.List(context.Background(), ProviderOpenAI)
	if err == nil {
		t.Fatal("expected error for HTTP 401")
	}
	if !IsAuthError(err) {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestModelServiceListEmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"object": "list",
			"data":   []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("test-key")
	cfg.baseURLs[ProviderOpenAI] = server.URL
	svc := newModelService(cfg)

	result, err := svc.List(context.Background(), ProviderOpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 0 {
		t.Errorf("expected 0 models, got %d", len(result.Models))
	}
}

func TestModelServiceListAuthHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer my-api-key" {
			t.Errorf("expected Authorization 'Bearer my-api-key', got %q", got)
		}
		resp := map[string]any{"data": []map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := defaultConfig("my-api-key")
	cfg.baseURLs[ProviderAnthropic] = server.URL
	svc := newModelService(cfg)

	_, err := svc.List(context.Background(), ProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModelServiceViaClient(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{
					"id":                "anthropic/claude-haiku-3.5",
					"display_name":      "Claude Haiku 3.5",
					"context_length":    200000,
					"input_modalities":  []string{"text"},
					"output_modalities": []string{"text"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key",
		WithBaseURL(ProviderAnthropic, server.URL),
	)

	if client.Models == nil {
		t.Fatal("expected Models service to be initialized")
	}

	result, err := client.Models.List(context.Background(), ProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Models))
	}
	if result.Models[0].ID != "anthropic/claude-haiku-3.5" {
		t.Errorf("unexpected ID: %s", result.Models[0].ID)
	}
}
