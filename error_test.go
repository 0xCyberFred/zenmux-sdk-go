package zenmux

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	e := &Error{
		Provider:   ProviderOpenAI,
		StatusCode: 429,
		Code:       "rate_limit_exceeded",
		Message:    "too many requests",
	}
	want := "zenmux openai error (HTTP 429, rate_limit_exceeded): too many requests"
	if got := e.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorMessageWithoutCode(t *testing.T) {
	e := &Error{
		Provider:   ProviderAnthropic,
		StatusCode: 500,
		Message:    "internal server error",
	}
	want := "zenmux anthropic error (HTTP 500): internal server error"
	if got := e.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("connection refused")
	e := &Error{Provider: ProviderOpenAI, StatusCode: 0, Message: "connection failed", Err: inner}
	if !errors.Is(e, inner) {
		t.Error("Unwrap should expose inner error")
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"429", &Error{StatusCode: 429}, true},
		{"422", &Error{StatusCode: 422}, true},
		{"500", &Error{StatusCode: 500}, false},
		{"non-zenmux", errors.New("other"), false},
		{"wrapped 429", fmt.Errorf("wrap: %w", &Error{StatusCode: 429}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimitError(tt.err); got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"401", &Error{StatusCode: 401}, true},
		{"403", &Error{StatusCode: 403}, true},
		{"200", &Error{StatusCode: 200}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	e := &Error{StatusCode: 404}
	if !IsNotFoundError(e) {
		t.Error("expected true for 404")
	}
	if IsNotFoundError(&Error{StatusCode: 200}) {
		t.Error("expected false for 200")
	}
}
