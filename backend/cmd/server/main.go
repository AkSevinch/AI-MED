package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"ai-med/backend/internal/ai"
	"ai-med/backend/internal/app"
	"ai-med/backend/internal/config"
	"ai-med/backend/internal/db"
	"ai-med/backend/internal/storage"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("database migrated")

	store, err := storage.NewLocal(cfg.UploadDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	provider := ai.New(ai.Options{
		Provider:       cfg.AIProvider,
		GeminiKey:      cfg.Gemini.APIKey,
		GeminiModel:    cfg.Gemini.Model,
		AnthropicKey:   cfg.Anthropic.APIKey,
		AnthropicModel: cfg.Anthropic.Model,
		OpenAIKey:      cfg.OpenAI.APIKey,
		OpenAIModel:    cfg.OpenAI.Model,
		OpenAIBaseURL:  cfg.OpenAI.BaseURL,
	})

	a := app.New(cfg, database, store, provider)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: a.Router(),
	}

	log.Printf("AI MED backend listening on :%s (ai provider: %s)", cfg.Port, provider.Name())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
