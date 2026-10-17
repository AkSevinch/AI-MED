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

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models"

// geminiMaxAttempts — число попыток при временных сбоях (503/429/5xx)
const geminiMaxAttempts = 4

// Gemini — провайдер ИИ на базе Google Gemini REST API
type Gemini struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGemini создаёт провайдер Gemini
func NewGemini(apiKey, model string) *Gemini {
	return &Gemini{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *Gemini) Name() string { return "gemini" }

type geminiPart struct {
	Text       string        `json:"text,omitempty"`
	InlineData *geminiInline `json:"inline_data,omitempty"`
}

type geminiInline struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"system_instruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  map[string]interface{} `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *Gemini) generate(ctx context.Context, req *geminiRequest) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("gemini: GEMINI_API_KEY is not set")
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiEndpoint, g.model, g.apiKey)

	var lastErr error
	backoff := time.Second
	for attempt := 1; attempt <= geminiMaxAttempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return "", err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := g.client.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("gemini request: %w", err)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		_ = resp.Body.Close()

		// временные сбои (перегрузка, лимиты) — повторяем с паузой
		if isRetryableStatus(resp.StatusCode) && attempt < geminiMaxAttempts {
			lastErr = fmt.Errorf("gemini status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
			continue
		}

		var parsed geminiResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return "", fmt.Errorf("gemini decode (status %d): %w", resp.StatusCode, err)
		}
		if parsed.Error != nil {
			return "", fmt.Errorf("gemini error: %s", parsed.Error.Message)
		}
		if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("gemini: empty response")
		}
		return parsed.Candidates[0].Content.Parts[0].Text, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("gemini: request failed")
	}
	return "", lastErr
}

// isRetryableStatus — коды, при которых запрос имеет смысл повторить
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

func extractionSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"lab_name":  map[string]interface{}{"type": "string"},
			"test_date": map[string]interface{}{"type": "string"},
			"indicators": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code":            map[string]interface{}{"type": "string"},
						"name":            map[string]interface{}{"type": "string"},
						"value":           map[string]interface{}{"type": "string"},
						"unit":            map[string]interface{}{"type": "string"},
						"reference_range": map[string]interface{}{"type": "string"},
						"test_date":       map[string]interface{}{"type": "string"},
						"confidence":      map[string]interface{}{"type": "number"},
					},
					"required": []string{"value"},
				},
			},
		},
		"required": []string{"indicators"},
	}
}

// Extract структурирует лабораторный документ
func (g *Gemini) Extract(ctx context.Context, c Content) (*ExtractionResult, error) {
	parts := []geminiPart{{
		Text: ExtractionSystemPrompt + "\n\nВерни результат строго по JSON-схеме.",
	}}
	if len(c.Images) > 0 {
		for _, img := range c.Images {
			mime := img.MIME
			if mime == "" {
				mime = "image/png"
			}
			parts = append(parts, geminiPart{InlineData: &geminiInline{
				MIMEType: mime,
				Data:     base64.StdEncoding.EncodeToString(img.Data),
			}})
		}
	} else if len(c.FileBytes) > 0 {
		mime := c.MIME
		if mime == "" {
			mime = "application/octet-stream"
		}
		parts = append(parts, geminiPart{InlineData: &geminiInline{
			MIMEType: mime,
			Data:     base64.StdEncoding.EncodeToString(c.FileBytes),
		}})
	} else if strings.TrimSpace(c.Text) != "" {
		parts = append(parts, geminiPart{Text: "ТЕКСТ ДОКУМЕНТА:\n" + c.Text})
	} else {
		return nil, fmt.Errorf("empty content")
	}

	req := &geminiRequest{
		Contents: []geminiContent{{Role: "user", Parts: parts}},
		GenerationConfig: map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema":   extractionSchema(),
			"temperature":      0,
		},
	}
	out, err := g.generate(ctx, req)
	if err != nil {
		return nil, err
	}
	return parseExtraction(out)
}

// Chat отвечает на вопрос по данным
func (g *Gemini) Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error) {
	contents := make([]geminiContent, 0, len(history)+1)
	for _, m := range history {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}
	contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: question}}})

	req := &geminiRequest{
		SystemInstruction: &geminiContent{Role: "system", Parts: []geminiPart{{Text: system}}},
		Contents:          contents,
	}
	return g.generate(ctx, req)
}
