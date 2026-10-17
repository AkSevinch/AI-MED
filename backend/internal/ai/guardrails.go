package ai

import (
	"regexp"
	"strings"
)

// ExtractionSystemPrompt — модель только структурирует данные
// без интерпретации, диагноза, лечения и выдуманных показателей
const ExtractionSystemPrompt = `Ты — модуль структурирования лабораторных данных для сервиса ведения личной сводки здоровья.
Твоя единственная задача — извлечь значения из документа.

СТРОГИЕ ПРАВИЛА:
1. Извлекай только показатели из поддерживаемого списка: общий холестерин, LDL (ЛПНП), HDL (ЛПВП),
   триглицериды, глюкоза, HbA1c, креатинин, калий, натрий.
2. НЕ интерпретируй значения. НЕ сравнивай с нормой. НЕ пиши статусы, диагнозы, приоритеты.
3. НЕ назначай лечение и НЕ упоминай лекарства или дозировки.
4. НЕ придумывай показатели, которых нет в документе. Если показатель отсутствует — не включай его.
5. Для каждого показателя верни: code, name, value, unit, reference_range (если указан), test_date
   (если указана), confidence (0..1 — насколько уверенно распознано).
6. Если единицы измерения не указаны — оставь пустую строку, не додумывай.
7. Ответ — строго JSON без пояснений, ровно такого вида (ключ со списком — обязательно "indicators"):
   {"lab_name": "", "test_date": "ГГГГ-ММ-ДД", "indicators": [
     {"code": "", "name": "", "value": "", "unit": "", "reference_range": "", "test_date": "", "confidence": 0.0}]}`

// ChatSystemPrompt ограничивает чат-ассистента (NFR9/NFR10)
const ChatSystemPrompt = `Ты — ассистент сервиса AI MED. Ты помогаешь пользователю понять его личную сводку здоровья.
Отвечай только на основе переданных подтверждённых данных пользователя.

СТРОГИЕ ЗАПРЕТЫ:
- Не назначать и не отменять препараты, не указывать дозировки.
- Не ставить диагноз и не выдавать предположение за установленный факт.
- Не придумывать показатели, значения и даты, которых нет в данных.
- Не использовать непроверенные данные.

СТИЛЬ:
- Если используешь показатель — назови показатель, значение, единицу и дату исследования.
- Если данных недостаточно — прямо сообщи об этом.
- При описании тревожных симптомов порекомендуй приоритетную очную или неотложную помощь (103/112).
- Отвечай на языке пользователя, кратко и понятно.`

// OverviewSystemPrompt — короткий обзор сводки и подготовка к визиту
// без диагноза, лечения и дозировок (FR8, FR15, US-15)
const OverviewSystemPrompt = `Ты — модуль подготовки к визиту врача в сервисе AI MED.
По переданным ПОДТВЕРЖДЁННЫМ данным пользователя составь короткий обзор.

СТРУКТУРА ОТВЕТА (строго три раздела):
1. «Кратко» — 2–4 предложения: что уже собрано в сводке.
2. «Отклонения от референса» — показатели со статусом «выше» или «ниже»: название, значение,
   единица и дата. Если отклонений нет — прямо напиши об этом.
3. «Что обсудить с врачом» — 3–5 вопросов врачу на основе этих данных.

СТРОГИЕ ЗАПРЕТЫ:
- Не ставить диагноз и не выдавать предположение за установленный факт.
- Не назначать и не отменять препараты, не указывать дозировки и схемы лечения.
- Не придумывать показатели, значения и даты, которых нет в данных.
- Не давать медицинских рекомендаций по лечению.

СТИЛЬ:
- Кратко, понятно, на русском языке.
- Если данных недостаточно — прямо сообщи об этом.
- При тревожных симптомах — приоритетная очная или неотложная помощь (103/112).`

// CheckOverviewAnswer возвращает безопасную замену при нарушении правил
func CheckOverviewAnswer(answer string) string {
	if prohibitedChat.MatchString(answer) {
		return SafeChatFallback
	}
	return answer
}

// supportedByName сопоставляет подстроки в нижнем регистре с кодами MVP
var supportedByName = map[string]string{
	"общий холестерин":  CodeTotalCholesterol,
	"холестерин общий":  CodeTotalCholesterol,
	"total cholesterol": CodeTotalCholesterol,
	"лпнп":              CodeLDL,
	"ldl":               CodeLDL,
	"лпвп":              CodeHDL,
	"hdl":               CodeHDL,
	"триглицерид":       CodeTriglycerides,
	"triglycerid":       CodeTriglycerides,
	"глюкоз":            CodeGlucose,
	"glucose":           CodeGlucose,
	"hba1c":             CodeHbA1c,
	"гликированн":       CodeHbA1c,
	"креатинин":         CodeCreatinine,
	"creatinine":        CodeCreatinine,
	"калий":             CodePotassium,
	"potassium":         CodePotassium,
	"натрий":            CodeSodium,
	"sodium":            CodeSodium,
}

func normalizeCode(ind ExtractedIndicator) string {
	if code, ok := SupportedIndicators[strings.ToLower(strings.TrimSpace(ind.Code))]; ok {
		return code
	}
	low := strings.ToLower(ind.Name)
	for needle, code := range supportedByName {
		if strings.Contains(low, needle) {
			return code
		}
	}
	return ""
}

// ApplyExtractionGuardrails фильтрует и нормализует результат извлечения
// неизвестные показатели отбрасываются, поля очищаются
func ApplyExtractionGuardrails(res *ExtractionResult) *ExtractionResult {
	if res == nil {
		return &ExtractionResult{}
	}
	out := &ExtractionResult{
		LabName:  strings.TrimSpace(res.LabName),
		TestDate: strings.TrimSpace(res.TestDate),
	}
	for _, ind := range res.Indicators {
		code := normalizeCode(ind)
		if code == "" {
			continue // нет в списке MVP
		}
		if strings.TrimSpace(ind.Value) == "" {
			continue
		}
		ind.Code = code
		if strings.TrimSpace(ind.Name) == "" {
			ind.Name = SupportedIndicators[code]
		}
		if ind.Confidence < 0 {
			ind.Confidence = 0
		}
		if ind.Confidence > 1 {
			ind.Confidence = 1
		}
		out.Indicators = append(out.Indicators, ind)
	}
	return out
}

var prohibitedChat = regexp.MustCompile(`(?i)(дозиров|принимайте|назнач|пропиш|отмени|замените препарат|мг в сутки|таблетк|инъекц|рекомендуется принимать|следует принимать|рекомендую принимать|увеличьте дозу|уменьшите дозу|снизьте дозу)`)

// SafeChatFallback возвращается при нарушении правил безопасности
const SafeChatFallback = "Я не могу назначать или отменять лечение, ставить диагноз или указывать дозировки. " +
	"Я могу помочь понять ваши подтверждённые данные сводки. При тревожных симптомах, пожалуйста, " +
	"немедленно обратитесь за медицинской помощью (103/112)."

// CheckChatAnswer возвращает безопасную замену при нарушении правил
// иначе исходный ответ
func CheckChatAnswer(answer string) string {
	if prohibitedChat.MatchString(answer) {
		return SafeChatFallback
	}
	return answer
}
