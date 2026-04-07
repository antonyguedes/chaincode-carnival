package providers

import (
	"fmt"
	"os"
	"strings"
)

// Config represents the raw ENV breakdown for an agent
type Config struct {
	Provider string
	Model    string
	APIKey   string
}

// NewProviderFromEnv reads dynamically from the .env environment prefix
// Examples: prefix="REDTEAM", "ANALYZER"
func NewProviderFromEnv(prefix string) (LLMProvider, error) {
	providerEngine := strings.ToLower(os.Getenv(fmt.Sprintf("%s_PROVIDER", prefix)))
	model := os.Getenv(fmt.Sprintf("%s_MODEL", prefix))
	apiKey := os.Getenv(fmt.Sprintf("%s_API_KEY", prefix))

	if providerEngine == "" {
		return nil, fmt.Errorf("missing %s_PROVIDER in .env", prefix)
	}

	switch providerEngine {
	case "ollama":
		return NewOllamaClient(model), nil
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("%s missing %s_API_KEY for Gemini", prefix, prefix)
		}
		return NewGeminiClient(model, apiKey), nil
	case "groq":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for Groq", prefix, prefix) }
		return NewOpenAICompatibleClient("Groq", "https://api.groq.com/openai/v1/chat/completions", model, apiKey), nil
	case "openai":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for OpenAI", prefix, prefix) }
		return NewOpenAICompatibleClient("OpenAI", "https://api.openai.com/v1/chat/completions", model, apiKey), nil
	case "deepseek":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for Deepseek", prefix, prefix) }
		return NewOpenAICompatibleClient("Deepseek", "https://api.deepseek.com/chat/completions", model, apiKey), nil
	case "mistralai":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for Mistral", prefix, prefix) }
		return NewOpenAICompatibleClient("MistralAI", "https://api.mistral.ai/v1/chat/completions", model, apiKey), nil
	case "openrouter":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for OpenRouter", prefix, prefix) }
		return NewOpenAICompatibleClient("OpenRouter", "https://openrouter.ai/api/v1/chat/completions", model, apiKey), nil
	case "anthropic":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for Anthropic", prefix, prefix) }
		return NewAnthropicClient(model, apiKey), nil
	case "cohere":
		if apiKey == "" { return nil, fmt.Errorf("%s missing %s_API_KEY for Cohere", prefix, prefix) }
		return NewCohereClient(model, apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider '%s' requested by %s", providerEngine, prefix)
	}
}
