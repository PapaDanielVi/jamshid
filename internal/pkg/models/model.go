package models

import "strings"

// Model represents an AI model.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
}

// AnthropicModels lists official Anthropic Claude models.
var AnthropicModels = []Model{
	{ID: "claude-opus-4-7", Name: "Claude Opus 4.7", Provider: "anthropic", Description: "Most capable model for complex tasks"},
	{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", Provider: "anthropic", Description: "Balanced performance and speed"},
	{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5", Provider: "anthropic", Description: "Fastest and most compact model"},
}

// OpenRouterModels lists common models available via OpenRouter.
var OpenRouterModels = []Model{
	{ID: "openrouter/anthropic/claude-opus-4-7", Name: "OpenRouter: Claude Opus 4.7", Provider: "openrouter", Description: "Via OpenRouter"},
	{ID: "openrouter/anthropic/claude-sonnet-4-6", Name: "OpenRouter: Claude Sonnet 4.6", Provider: "openrouter", Description: "Via OpenRouter"},
	{ID: "openrouter/anthropic/claude-haiku-4-5", Name: "OpenRouter: Claude Haiku 4.5", Provider: "openrouter", Description: "Via OpenRouter"},
	{ID: "openrouter/google/gemini-2.5-pro", Name: "OpenRouter: Gemini 2.5 Pro", Provider: "openrouter", Description: "Via OpenRouter"},
	{ID: "openrouter/meta-llama/llama-4-maverick", Name: "OpenRouter: Llama 4 Maverick", Provider: "openrouter", Description: "Via OpenRouter"},
}

// AllModels returns all available models.
func AllModels() []Model {
	return append(append([]Model{}, AnthropicModels...), OpenRouterModels...)
}

// SearchModels returns models matching the query (case-insensitive substring match on ID and Name).
func SearchModels(query string) []Model {
	if query == "" {
		return AllModels()
	}
	query = strings.ToLower(query)
	var results []Model
	for _, m := range AllModels() {
		if strings.Contains(strings.ToLower(m.ID), query) || strings.Contains(strings.ToLower(m.Name), query) {
			results = append(results, m)
		}
	}
	return results
}
