package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// openAIModelAttempts — попыток на одну модель при временных сбоях (429/5xx)
const openAIModelAttempts = 2

// OpenAI — провайдер для любого OpenAI-совместимого API
// (OpenRouter, Groq, GitHub Models, Cloudflare Workers AI, Together и т.п.)
type OpenAI struct {
	apiKey  string
	baseURL string
	models  []string
	name    string
	client  *http.Client
}

// NewOpenAI создаёт OpenAI-совместимый провайдер.
// model может быть списком через запятую — модели перебираются по порядку
// (удобно, когда бесплатные модели упираются в общий upstream-лимит).
func NewOpenAI(apiKey, model, baseURL string) *OpenAI {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	name := "openai"
	if strings.Contains(baseURL, "openrouter") {
		name = "openrouter"
	}
	models := make([]string, 0, 4)
	for _, p := range strings.Split(model, ",") {
		if p = strings.TrimSpace(p); p != "" {
			models = append(models, p)
		}
	}
	if len(models) == 0 {
		models = []string{"google/gemma-4-31b-it:free"}
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: baseURL,
		models:  models,
		name:    name,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (o *OpenAI) Name() string { return o.name }

type openAIContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *openAIImage `json:"image_url,omitempty"`
}

type openAIImage struct {
	URL string `json:"url"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type openAIRequest struct {
	Model          string          `json:"model"`
	Messages       []openAIMessage `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat map[string]any  `json:"response_format,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message  string `json:"message"`
		Code     any    `json:"code"`
		Metadata struct {
			Raw string `json:"raw"`
		} `json:"metadata"`
	} `json:"error"`
}

// contentText достаёт текст из content (строка или массив частей)
func contentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var sb strings.Builder
		for _, p := range parts {
			if p.Type == "text" || p.Text != "" {
				sb.WriteString(p.Text)
			}
		}
		return sb.String()
	}
	return ""
}

// imagePart формирует часть контента с изображением в виде data-URL
func imagePart(data []byte, mime string) openAIContentPart {
	if mime == "" {
		mime = "image/png"
	}
	url := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
	return openAIContentPart{Type: "image_url", ImageURL: &openAIImage{URL: url}}
}

// callPost выполняет запрос к одной модели с повторами при временных сбоях
func (o *OpenAI) callPost(ctx context.Context, url, model string, raw []byte) (string, error) {
	var lastErr error
	backoff := time.Second
	for attempt := 1; attempt <= openAIModelAttempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return "", err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
		// необязательные заголовки OpenRouter
		httpReq.Header.Set("HTTP-Referer", "https://github.com/ai-med")
		httpReq.Header.Set("X-Title", "AI MED")

		resp, err := o.client.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("%s[%s] request: %w", o.name, model, err)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		_ = resp.Body.Close()

		if isRetryableStatus(resp.StatusCode) && attempt < openAIModelAttempts {
			lastErr = fmt.Errorf("%s[%s] status %d: %s", o.name, model, resp.StatusCode, strings.TrimSpace(string(body)))
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
			continue
		}

		var parsed openAIResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return "", fmt.Errorf("%s[%s] decode (status %d): %w", o.name, model, resp.StatusCode, err)
		}
		if parsed.Error != nil {
			msg := parsed.Error.Message
			if parsed.Error.Metadata.Raw != "" {
				msg = parsed.Error.Metadata.Raw
			}
			return "", fmt.Errorf("%s[%s] error (status %d): %s", o.name, model, resp.StatusCode, msg)
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("%s[%s]: empty response (status %d): %s", o.name, model, resp.StatusCode, strings.TrimSpace(string(body)))
		}
		out := contentText(parsed.Choices[0].Message.Content)
		if strings.TrimSpace(out) == "" {
			return "", fmt.Errorf("%s[%s]: empty message", o.name, model)
		}
		return out, nil
	}
	return "", lastErr
}

// call перебирает модели по очереди, пока одна не ответит
func (o *OpenAI) call(ctx context.Context, req *openAIRequest) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("%s: API key is not set", o.name)
	}
	if o.baseURL == "" {
		return "", fmt.Errorf("%s: base URL is not set", o.name)
	}
	url := o.baseURL + "/chat/completions"

	var errs []string
	for _, model := range o.models {
		req.Model = model
		raw, err := json.Marshal(req)
		if err != nil {
			return "", err
		}
		out, err := o.callPost(ctx, url, model, raw)
		if err == nil {
			return out, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return "", fmt.Errorf("%s: request failed", o.name)
	}
	return "", fmt.Errorf("%s: все модели недоступны: %s", o.name, strings.Join(errs, " | "))
}

// Extract структурирует лабораторный документ
func (o *OpenAI) Extract(ctx context.Context, c Content) (*ExtractionResult, error) {
	parts := []openAIContentPart{{
		Type: "text",
		Text: ExtractionSystemPrompt + "\n\nВерни результат строго как JSON-объект.",
	}}

	switch {
	case len(c.Images) > 0:
		for _, img := range c.Images {
			parts = append(parts, imagePart(img.Data, img.MIME))
		}
	case len(c.FileBytes) > 0:
		parts = append(parts, imagePart(c.FileBytes, c.MIME))
	case strings.TrimSpace(c.Text) != "":
		parts = append(parts, openAIContentPart{Type: "text", Text: "ТЕКСТ ДОКУМЕНТА:\n" + c.Text})
	default:
		return nil, fmt.Errorf("empty content")
	}

	zero := 0.0
	out, err := o.call(ctx, &openAIRequest{
		Messages:       []openAIMessage{{Role: "user", Content: parts}},
		Temperature:    &zero,
		MaxTokens:      8192,
		ResponseFormat: map[string]any{"type": "json_object"},
	})
	if err != nil {
		return nil, err
	}
	return parseExtraction(out)
}

// Chat отвечает на вопрос по данным
func (o *OpenAI) Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error) {
	msgs := make([]openAIMessage, 0, len(history)+2)
	if strings.TrimSpace(system) != "" {
		msgs = append(msgs, openAIMessage{Role: "system", Content: system})
	}
	for _, m := range history {
		role := m.Role
		if role != "assistant" {
			role = "user"
		}
		msgs = append(msgs, openAIMessage{Role: role, Content: m.Content})
	}
	msgs = append(msgs, openAIMessage{Role: "user", Content: question})

	return o.call(ctx, &openAIRequest{Messages: msgs, MaxTokens: 2048})
}
