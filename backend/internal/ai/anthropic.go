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

const anthropicEndpoint = "https://api.anthropic.com/v1/messages"

// Anthropic — провайдер ИИ на базе Claude Messages API
type Anthropic struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropic создаёт провайдер Anthropic
func NewAnthropic(apiKey, model string) *Anthropic {
	return &Anthropic{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (a *Anthropic) Name() string { return "anthropic" }

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Anthropic) call(ctx context.Context, req *anthropicRequest) (string, error) {
	if a.apiKey == "" {
		return "", fmt.Errorf("anthropic: ANTHROPIC_API_KEY is not set")
	}
	req.Model = a.model
	raw, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))

	var parsed anthropicResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("anthropic decode (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", parsed.Error.Message)
	}
	var sb strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return sb.String(), nil
}

// Extract структурирует лабораторный документ
func (a *Anthropic) Extract(ctx context.Context, c Content) (*ExtractionResult, error) {
	blocks := []map[string]any{{
		"type": "text",
		"text": ExtractionSystemPrompt + "\n\nВерни результат строго как JSON-объект.",
	}}
	if len(c.Images) > 0 {
		for _, img := range c.Images {
			mime := img.MIME
			if mime == "" {
				mime = "image/png"
			}
			blocks = append(blocks, map[string]any{
				"type": "image",
				"source": map[string]any{
					"type":       "base64",
					"media_type": mime,
					"data":       base64.StdEncoding.EncodeToString(img.Data),
				},
			})
		}
	} else if len(c.FileBytes) > 0 {
		if c.MIME == "application/pdf" {
			blocks = append(blocks, map[string]any{
				"type": "document",
				"source": map[string]any{
					"type":       "base64",
					"media_type": c.MIME,
					"data":       base64.StdEncoding.EncodeToString(c.FileBytes),
				},
			})
		} else {
			blocks = append(blocks, map[string]any{
				"type": "image",
				"source": map[string]any{
					"type":       "base64",
					"media_type": c.MIME,
					"data":       base64.StdEncoding.EncodeToString(c.FileBytes),
				},
			})
		}
	} else if strings.TrimSpace(c.Text) != "" {
		blocks = append(blocks, map[string]any{"type": "text", "text": "ТЕКСТ ДОКУМЕНТА:\n" + c.Text})
	} else {
		return nil, fmt.Errorf("empty content")
	}

	out, err := a.call(ctx, &anthropicRequest{
		MaxTokens: 4096,
		Messages:  []anthropicMessage{{Role: "user", Content: blocks}},
	})
	if err != nil {
		return nil, err
	}
	return parseExtraction(out)
}

// Chat отвечает на вопрос по данным
func (a *Anthropic) Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error) {
	msgs := make([]anthropicMessage, 0, len(history)+1)
	for _, m := range history {
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}
	msgs = append(msgs, anthropicMessage{Role: "user", Content: question})
	return a.call(ctx, &anthropicRequest{MaxTokens: 1500, System: system, Messages: msgs})
}

// extractJSON достаёт первый JSON-объект из ответа
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "json")
		if j := strings.Index(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
