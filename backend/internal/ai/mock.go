package ai

import (
	"context"
	"strings"
	"time"
)

// Mock — заглушка провайдера ИИ для демонстрации без внешних ключей
type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Name() string { return "mock" }

// Extract возвращает правдоподобные показатели MVP (демо-данные)
func (m *Mock) Extract(ctx context.Context, c Content) (*ExtractionResult, error) {
	today := time.Now().Format("2006-01-02")
	ind := func(code, value, unit, ref string) ExtractedIndicator {
		return ExtractedIndicator{
			Code:           code,
			Name:           SupportedIndicators[code],
			Value:          value,
			Unit:           unit,
			ReferenceRange: ref,
			TestDate:       today,
			Confidence:     0.92,
		}
	}
	return &ExtractionResult{
		LabName:  "Городская медицинская лаборатория «ДИАГНОСТИКА+»",
		TestDate: today,
		Indicators: []ExtractedIndicator{
			ind(CodeTotalCholesterol, "5,4", "ммоль/л", "<5,2"),
			ind(CodeLDL, "3,6", "ммоль/л", "<3,0"),
			ind(CodeHDL, "1,2", "ммоль/л", ">1,0"),
			ind(CodeTriglycerides, "1,8", "ммоль/л", "<1,7"),
			ind(CodeGlucose, "5,1", "ммоль/л", "3,9–5,6"),
			ind(CodeHbA1c, "5,6", "%", "<6,0"),
			ind(CodeCreatinine, "78", "мкмоль/л", "58–96"),
			ind(CodePotassium, "4,3", "ммоль/л", "3,5–5,1"),
			ind(CodeSodium, "140", "ммоль/л", "136–145"),
		},
	}, nil
}

// Chat отвечает демо-текстом в рамках guardrails
func (m *Mock) Chat(ctx context.Context, system, question string, history []ChatMessage) (string, error) {
	var sb strings.Builder
	sb.WriteString("Демонстрационный режим: внешний ИИ не подключён, ответ собран из ваших данных в сводке.\n\n")
	sb.WriteString("В сводке два показателя выше референса — ЛПНП (3,6 ммоль/л при <3,0) и общий холестерин (5,4 ммоль/л при <5,2). Последнее измеренное давление — 140/90 мм рт. ст.\n\n")
	sb.WriteString("Что можно обсудить с врачом:\n")
	sb.WriteString("- стоит ли повторить липидограмму и как часто её контролировать;\n")
	sb.WriteString("- нужны ли изменения в питании и физической активности;\n")
	sb.WriteString("- как правильно измерять давление дома и какие значения записывать.\n\n")
	sb.WriteString("Это информационная подсказка, а не медицинское заключение; окончательное решение принимает врач.")
	return sb.String(), nil
}
