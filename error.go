package zenmux

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
)

type Error struct {
	Provider   Provider
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("zenmux %s error (HTTP %d, %s): %s", e.Provider, e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("zenmux %s error (HTTP %d): %s", e.Provider, e.StatusCode, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func IsRateLimitError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusTooManyRequests || e.StatusCode == 422
	}
	return false
}

func IsAuthError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
	}
	return false
}

func IsNotFoundError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

// wrapOpenAIError converts an error from the OpenAI SDK into a zenmux Error,
// extracting the HTTP status code when available.
func wrapOpenAIError(err error) error {
	if err == nil {
		return nil
	}
	e := &Error{
		Provider: ProviderOpenAI,
		Message:  err.Error(),
		Err:      err,
	}
	var apierr *openai.Error
	if errors.As(err, &apierr) {
		e.StatusCode = apierr.StatusCode
	}
	return e
}

// wrapAnthropicError converts an error from the Anthropic SDK into a zenmux
// Error, extracting the HTTP status code when available.
func wrapAnthropicError(err error) error {
	if err == nil {
		return nil
	}
	e := &Error{
		Provider: ProviderAnthropic,
		Message:  err.Error(),
		Err:      err,
	}
	var apierr *anthropic.Error
	if errors.As(err, &apierr) {
		e.StatusCode = apierr.StatusCode
	}
	return e
}
