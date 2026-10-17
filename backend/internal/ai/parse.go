package ai

import (
	"encoding/json"
	"strings"
)

// parseExtraction разбирает JSON-ответ модели в ExtractionResult.
// Принимает ключ "indicators", а также распространённые синонимы
// ("results", "data", "items") и «голый» массив показателей.
func parseExtraction(s string) (*ExtractionResult, error) {
	clean := extractJSON(s)

	// «голый» массив: [ {...}, {...} ]
	if strings.HasPrefix(strings.TrimSpace(clean), "[") {
		var arr []ExtractedIndicator
		if err := json.Unmarshal([]byte(clean), &arr); err != nil {
			return nil, err
		}
		return &ExtractionResult{Indicators: arr}, nil
	}

	var res ExtractionResult
	if err := json.Unmarshal([]byte(clean), &res); err != nil {
		return nil, err
	}
	if len(res.Indicators) > 0 {
		return &res, nil
	}

	// синонимы верхнего уровня
	var alt struct {
		LabName    string               `json:"lab_name"`
		TestDate   string               `json:"test_date"`
		Results    []ExtractedIndicator `json:"results"`
		Data       []ExtractedIndicator `json:"data"`
		Items      []ExtractedIndicator `json:"items"`
		Indicators []ExtractedIndicator `json:"indicators"`
	}
	if err := json.Unmarshal([]byte(clean), &alt); err != nil {
		return &res, nil
	}
	res.Indicators = firstNonEmptyList(alt.Indicators, alt.Results, alt.Data, alt.Items)
	if strings.TrimSpace(res.LabName) == "" {
		res.LabName = alt.LabName
	}
	if strings.TrimSpace(res.TestDate) == "" {
		res.TestDate = alt.TestDate
	}
	return &res, nil
}

// firstNonEmptyList возвращает первый непустой список
func firstNonEmptyList(lists ...[]ExtractedIndicator) []ExtractedIndicator {
	for _, l := range lists {
		if len(l) > 0 {
			return l
		}
	}
	return nil
}
