package ai

import (
	"context"
	"strings"
)

// Коды показателей MVP (SRS FR6)
const (
	CodeTotalCholesterol = "total_cholesterol"
	CodeLDL              = "ldl"
	CodeHDL              = "hdl"
	CodeTriglycerides    = "triglycerides"
	CodeGlucose          = "glucose"
	CodeHbA1c            = "hba1c"
	CodeCreatinine       = "creatinine"
	CodePotassium        = "potassium"
	CodeSodium           = "sodium"
)

// SupportedIndicators сопоставляет коды MVP с русскими названиями
var SupportedIndicators = map[string]string{
	CodeTotalCholesterol: "Общий холестерин",
	CodeLDL:              "ЛПНП (LDL-холестерин)",
	CodeHDL:              "ЛПВП (HDL-холестерин)",
	CodeTriglycerides:    "Триглицериды",
	CodeGlucose:          "Глюкоза",
	CodeHbA1c:            "HbA1c (гликированный гемоглобин)",
	CodeCreatinine:       "Креатинин",
	CodePotassium:        "Калий",
	CodeSodium:           "Натрий",
}

// DefaultReference — типовые референсные диапазоны для показателей MVP.
// Используются, когда в бланке анализов референс не указан (FR8).
var DefaultReference = map[string]string{
	CodeTotalCholesterol: "<5,2",
	CodeLDL:              "<3,0",
	CodeHDL:              ">1,0",
	CodeTriglycerides:    "<1,7",
	CodeGlucose:          "3,9–5,6",
	CodeHbA1c:            "<6,0",
	CodeCreatinine:       "59–104",
	CodePotassium:        "3,5–5,1",
	CodeSodium:           "136–145",
}

// ReferenceFor возвращает указанный в бланке референс, а при его отсутствии —
// типовой диапазон для этого кода показателя
func ReferenceFor(code, extracted string) string {
	if s := strings.TrimSpace(extracted); s != "" {
		return s
	}
	return DefaultReference[code]
}

// ExtractedIndicator — одно лабораторное значение
// только поля из SRS: без статуса, приоритета, диагноза и рекомендаций
type ExtractedIndicator struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Unit           string  `json:"unit"`
	ReferenceRange string  `json:"reference_range"`
	TestDate       string  `json:"test_date"`
	Confidence     float64 `json:"confidence"`
}

// ExtractionResult — структурированный результат документа
type ExtractionResult struct {
	LabName    string               `json:"lab_name"`
	TestDate   string               `json:"test_date"`
	Indicators []ExtractedIndicator `json:"indicators"`
}

// Content — что анализирует провайдер: текст или бинарный
// документ (для провайдеров со зрением)
type Content struct {
	Text      string
	FileBytes []byte
	MIME      string
	Filename  string
	// Images — готовые изображения (например, страницы PDF,
	// растеризованные в PNG). Если заданы, используются вместо FileBytes.
	Images []ImagePart
}

// ImagePart — одно изображение для отправки провайдеру
type ImagePart struct {
	Data []byte
	MIME string
}

// ChatMessage — одна реплика в диалоге
type ChatMessage struct {
	Role    string // "user" | "assistant"
	Content string
}

// Provider — не зависящий от вендора интерфейс ИИ
type Provider interface {
	Name() string
	// Extract извлекает показатели MVP из документа
	Extract(ctx context.Context, c Content) (*ExtractionResult, error)
	// Chat отвечает на вопрос по переданным данным
	// не диагностирует и не назначает, guardrails применяет вызывающий
	Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error)
}
