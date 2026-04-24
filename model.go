package zenmux

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ModelList holds a collection of normalized models returned from a provider.
type ModelList struct {
	Models []Model
}

// Model is a unified representation of a model from any supported provider.
type Model struct {
	ID               string
	DisplayName      string
	Provider         Provider
	InputModalities  []string
	OutputModalities []string
	ContextLength    int
	Reasoning        bool
	Pricings         map[string][]Pricing
}

// Pricing describes a single pricing entry for a model.
type Pricing struct {
	Value      float64
	Unit       string
	Currency   string
	Conditions *PricingConditions
}

// PricingConditions specifies token-range conditions under which a pricing
// entry applies.
type PricingConditions struct {
	PromptTokens     *TokenRange `json:"prompt_tokens,omitempty"`
	CompletionTokens *TokenRange `json:"completion_tokens,omitempty"`
}

// TokenRange describes a numeric range for token-based pricing conditions.
type TokenRange struct {
	Gte *float64 `json:"gte,omitempty"`
	Lte *float64 `json:"lte,omitempty"`
	Gt  *float64 `json:"gt,omitempty"`
	Lt  *float64 `json:"lt,omitempty"`
}

// ModelService provides access to model listing across providers.
type ModelService struct {
	cfg *config
}

func newModelService(cfg *config) *ModelService {
	return &ModelService{cfg: cfg}
}

// List retrieves the available models for the given provider, returning a
// unified ModelList.
func (s *ModelService) List(ctx context.Context, provider Provider) (*ModelList, error) {
	switch provider {
	case ProviderOpenAI:
		return s.listFromEndpoint(ctx, s.cfg.baseURL(ProviderOpenAI)+"/models", ProviderOpenAI)
	case ProviderAnthropic:
		return s.listFromEndpoint(ctx, s.cfg.baseURL(ProviderAnthropic)+"/v1/models", ProviderAnthropic)
	case ProviderGoogle:
		return s.listFromEndpoint(ctx, s.cfg.baseURL(ProviderGoogle)+"/v1beta/models", ProviderGoogle)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// listFromEndpoint fetches the JSON model list from url, parses the
// provider-specific format, and normalises the result into a ModelList.
func (s *ModelService) listFromEndpoint(ctx context.Context, url string, provider Provider) (*ModelList, error) {
	hc := s.cfg.httpClient
	if hc == nil {
		hc = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.apiKey)

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &Error{
			Provider:   provider,
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	switch provider {
	case ProviderOpenAI:
		return parseOpenAIModels(body, provider)
	case ProviderAnthropic:
		return parseAnthropicModels(body, provider)
	case ProviderGoogle:
		return parseGoogleModels(body, provider)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// --- OpenAI response format ---

type openAIModelResponse struct {
	Data []openAIModel `json:"data"`
}

type openAIModel struct {
	ID               string              `json:"id"`
	DisplayName      string              `json:"display_name"`
	ContextLength    int                 `json:"context_length"`
	InputModalities  []string            `json:"input_modalities"`
	OutputModalities []string            `json:"output_modalities"`
	Capabilities     *modelCapabilities  `json:"capabilities"`
	Pricings         map[string][]rawPricing `json:"pricings"`
}

// --- Anthropic response format ---

type anthropicModelResponse struct {
	Data []anthropicModel `json:"data"`
}

type anthropicModel struct {
	ID               string              `json:"id"`
	DisplayName      string              `json:"display_name"`
	ContextLength    int                 `json:"context_length"`
	InputModalities  []string            `json:"input_modalities"`
	OutputModalities []string            `json:"output_modalities"`
	Capabilities     *modelCapabilities  `json:"capabilities"`
	Pricings         map[string][]rawPricing `json:"pricings"`
}

// --- Google response format ---

type googleModelResponse struct {
	Models []googleModel `json:"models"`
}

type googleModel struct {
	Name             string              `json:"name"`
	DisplayName      string              `json:"displayName"`
	InputTokenLimit  int                 `json:"inputTokenLimit"`
	InputModalities  []string            `json:"input_modalities"`
	OutputModalities []string            `json:"output_modalities"`
	Capabilities     *modelCapabilities  `json:"capabilities"`
	Pricings         map[string][]rawPricing `json:"pricings"`
}

// --- Shared raw types ---

type modelCapabilities struct {
	Reasoning bool `json:"reasoning"`
}

type rawPricing struct {
	Value      float64            `json:"value"`
	Unit       string             `json:"unit"`
	Currency   string             `json:"currency"`
	Conditions *PricingConditions `json:"conditions,omitempty"`
}

// --- Parsers ---

func parseOpenAIModels(body []byte, provider Provider) (*ModelList, error) {
	var resp openAIModelResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing OpenAI models: %w", err)
	}
	models := make([]Model, 0, len(resp.Data))
	for _, m := range resp.Data {
		models = append(models, Model{
			ID:               m.ID,
			DisplayName:      m.DisplayName,
			Provider:         provider,
			InputModalities:  m.InputModalities,
			OutputModalities: m.OutputModalities,
			ContextLength:    m.ContextLength,
			Reasoning:        m.Capabilities != nil && m.Capabilities.Reasoning,
			Pricings:         convertPricings(m.Pricings),
		})
	}
	return &ModelList{Models: models}, nil
}

func parseAnthropicModels(body []byte, provider Provider) (*ModelList, error) {
	var resp anthropicModelResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing Anthropic models: %w", err)
	}
	models := make([]Model, 0, len(resp.Data))
	for _, m := range resp.Data {
		models = append(models, Model{
			ID:               m.ID,
			DisplayName:      m.DisplayName,
			Provider:         provider,
			InputModalities:  m.InputModalities,
			OutputModalities: m.OutputModalities,
			ContextLength:    m.ContextLength,
			Reasoning:        m.Capabilities != nil && m.Capabilities.Reasoning,
			Pricings:         convertPricings(m.Pricings),
		})
	}
	return &ModelList{Models: models}, nil
}

func parseGoogleModels(body []byte, provider Provider) (*ModelList, error) {
	var resp googleModelResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing Google models: %w", err)
	}
	models := make([]Model, 0, len(resp.Models))
	for _, m := range resp.Models {
		models = append(models, Model{
			ID:               m.Name,
			DisplayName:      m.DisplayName,
			Provider:         provider,
			InputModalities:  m.InputModalities,
			OutputModalities: m.OutputModalities,
			ContextLength:    m.InputTokenLimit,
			Reasoning:        m.Capabilities != nil && m.Capabilities.Reasoning,
			Pricings:         convertPricings(m.Pricings),
		})
	}
	return &ModelList{Models: models}, nil
}

func convertPricings(raw map[string][]rawPricing) map[string][]Pricing {
	if raw == nil {
		return nil
	}
	result := make(map[string][]Pricing, len(raw))
	for k, rps := range raw {
		ps := make([]Pricing, 0, len(rps))
		for _, rp := range rps {
			ps = append(ps, Pricing{
				Value:      rp.Value,
				Unit:       rp.Unit,
				Currency:   rp.Currency,
				Conditions: rp.Conditions,
			})
		}
		result[k] = ps
	}
	return result
}
