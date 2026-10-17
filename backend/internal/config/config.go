package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config — вся конфигурация, читаемая из переменных окружения
type Config struct {
	Port          string
	DatabaseURL   string
	SessionSecret string
	UploadDir     string

	AIProvider string
	Gemini     ProviderConfig
	Anthropic  ProviderConfig
	OpenAI     ProviderConfig

	CORSOrigins []string
}

// ProviderConfig описывает одного провайдера ИИ
type ProviderConfig struct {
	APIKey  string
	Model   string
	BaseURL string // только для OpenAI-совместимых провайдеров
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// firstNonEmpty возвращает первое непустое значение
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// Load читает конфигурацию из окружения с dev-значениями по умолчанию
func Load() (*Config, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env("DB_USER", "postgres"),
			env("DB_PASSWORD", "password"),
			env("DB_HOST", "localhost"),
			env("DB_PORT", "5432"),
			env("DB_NAME", "medical_db"),
		)
	}

	// переопределение ssl-режима БД через DB_SSLMODE
	if ssl := strings.TrimSpace(os.Getenv("DB_SSLMODE")); ssl != "" && !strings.Contains(dsn, "sslmode=") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "sslmode=" + ssl
	}

	cfg := &Config{
		Port:          env("PORT", "5001"),
		DatabaseURL:   dsn,
		SessionSecret: env("SECRET_KEY", "dev-secret-key-change-me"),
		UploadDir:     env("UPLOAD_DIR", "uploads"),
		AIProvider:    strings.ToLower(env("AI_PROVIDER", "mock")),
		Gemini: ProviderConfig{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  env("GEMINI_MODEL", "gemini-3.8-flash"),
		},
		Anthropic: ProviderConfig{
			APIKey: os.Getenv("ANTHROPIC_API_KEY"),
			Model:  env("ANTHROPIC_MODEL", "claude-3-7-sonnet-latest"),
		},
		OpenAI: ProviderConfig{
			APIKey:  firstNonEmpty(os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENROUTER_API_KEY")),
			Model:   env("OPENAI_MODEL", "nvidia/nemotron-3.5-super-vl:free,nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free,dots-studio/dots-3-note-preview:free,google/gemma-4-31b-it:free"),
			BaseURL: env("OPENAI_BASE_URL", "https://openrouter.ai/api/v1"),
		},
		CORSOrigins: splitOrigins(env("CORS_ORIGINS",
			"http://localhost:3000,http://127.0.0.1:3000")),
	}
	return cfg, nil
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// AtoiDefault — хелпер для необязательных целочисленных переменных окружения
func AtoiDefault(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
