package zenmux

import (
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	client := NewClient("test-key")
	if client.Chat == nil {
		t.Error("Chat should not be nil")
	}
	if client.Responses == nil {
		t.Error("Responses should not be nil")
	}
	if client.Embeddings == nil {
		t.Error("Embeddings should not be nil")
	}
	if client.Messages == nil {
		t.Error("Messages should not be nil")
	}
	if client.Gemini == nil {
		t.Error("Gemini should not be nil")
	}
	if client.Models == nil {
		t.Error("Models should not be nil")
	}
	if client.Platform != nil {
		t.Error("Platform should be nil without management key")
	}
}

func TestNewClientWithManagementKey(t *testing.T) {
	client := NewClient("test-key", WithManagementKey("mgmt-key"))
	if client.Platform == nil {
		t.Error("Platform should not be nil with management key")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient("test-key",
		WithMaxRetries(5),
		WithTimeout(60*time.Second),
		WithBaseURL(ProviderOpenAI, "https://custom.example.com"),
	)
	if client.cfg.maxRetries != 5 {
		t.Errorf("expected maxRetries=5, got %d", client.cfg.maxRetries)
	}
	if client.cfg.timeout != 60*time.Second {
		t.Errorf("expected timeout=60s, got %v", client.cfg.timeout)
	}
	if client.cfg.baseURLs[ProviderOpenAI] != "https://custom.example.com" {
		t.Errorf("unexpected base URL: %s", client.cfg.baseURLs[ProviderOpenAI])
	}
}

func TestEscapeHatches(t *testing.T) {
	client := NewClient("test-key")
	// OpenAI and Anthropic are value types, always non-zero
	_ = client.OpenAI()
	_ = client.Anthropic()
	// Google is a pointer, may be nil if creation failed
}
