package ai

import (
	"context"
	"fmt"
	"strings"
)

// Fallback пробует основной провайдер, при ошибке — резервный
type Fallback struct {
	primary   Provider
	secondary Provider
}

func (f *Fallback) Name() string { return f.primary.Name() }

// Extract реализует Provider
func (f *Fallback) Extract(ctx context.Context, c Content) (*ExtractionResult, error) {
	res, err := f.primary.Extract(ctx, c)
	if err == nil {
		return res, nil
	}
	if f.secondary == nil {
		return nil, err
	}
	res2, err2 := f.secondary.Extract(ctx, c)
	if err2 != nil {
		return nil, fmt.Errorf("primary(%s): %v; fallback(%s): %w",
			f.primary.Name(), err, f.secondary.Name(), err2)
	}
	return res2, nil
}

// Chat реализует Provider
func (f *Fallback) Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error) {
	ans, err := f.primary.Chat(ctx, system, question, history)
	if err == nil {
		return ans, nil
	}
	if f.secondary == nil {
		return "", err
	}
	ans2, err2 := f.secondary.Chat(ctx, system, question, history)
	if err2 != nil {
		return "", fmt.Errorf("primary(%s): %v; fallback(%s): %w",
			f.primary.Name(), err, f.secondary.Name(), err2)
	}
	return ans2, nil
}

type unavailable struct{ reason string }

func (u *unavailable) Name() string { return "unavailable" }
func (u *unavailable) Extract(context.Context, Content) (*ExtractionResult, error) {
	return nil, fmt.Errorf("AI provider unavailable: %s", u.reason)
}
func (u *unavailable) Chat(context.Context, string, string, []ChatMessage) (string, error) {
	return "", fmt.Errorf("AI provider unavailable: %s", u.reason)
}

// Options — конфигурация выбора провайдера ИИ
type Options struct {
	Provider string

	GeminiKey   string
	GeminiModel string

	AnthropicKey   string
	AnthropicModel string

	OpenAIKey     string
	OpenAIModel   string
	OpenAIBaseURL string
}

// New выбирает провайдер по конфигурации и строит цепочку fallback:
// выбранный провайдер первый, затем остальные доступные в порядке
// openrouter → gemini → anthropic.
func New(opts Options) Provider {
	switch strings.ToLower(strings.TrimSpace(opts.Provider)) {
	case "mock", "demo":
		return NewMock()
	}

	byName := map[string]Provider{}
	if opts.OpenAIKey != "" {
		byName["openai"] = NewOpenAI(opts.OpenAIKey, opts.OpenAIModel, opts.OpenAIBaseURL)
	}
	if opts.GeminiKey != "" {
		byName["gemini"] = NewGemini(opts.GeminiKey, opts.GeminiModel)
	}
	if opts.AnthropicKey != "" {
		byName["anthropic"] = NewAnthropic(opts.AnthropicKey, opts.AnthropicModel)
	}

	ordered := make([]Provider, 0, len(byName))
	seen := map[string]bool{}
	add := func(name string) {
		if p, ok := byName[name]; ok && !seen[name] {
			ordered = append(ordered, p)
			seen[name] = true
		}
	}

	// выбранный провайдер — первым
	switch strings.ToLower(strings.TrimSpace(opts.Provider)) {
	case "anthropic", "claude":
		add("anthropic")
	case "gemini", "google":
		add("gemini")
	case "openai", "openrouter":
		add("openai")
	default:
		add("openai")
	}
	// резервные — в фиксированном порядке
	add("openai")
	add("gemini")
	add("anthropic")

	if len(ordered) == 0 {
		return NewMock()
	}
	// сворачиваем в цепочку fallback (primary → secondary → …)
	var result Provider
	for i := len(ordered) - 1; i >= 0; i-- {
		if result == nil {
			result = ordered[i]
			continue
		}
		result = &Fallback{primary: ordered[i], secondary: result}
	}
	return result
}
