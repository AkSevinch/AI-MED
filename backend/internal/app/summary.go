package app

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ai-med/backend/internal/ai"
	"ai-med/backend/internal/httpx"
)

// --- FR8: отклонения от референса ---

// indicatorView — подтверждённый показатель с вычисленным отклонением
type indicatorView struct {
	ID             int    `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Value          string `json:"value"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"reference_range"`
	TestDate       string `json:"test_date"`
	Source         string `json:"source"`
	Status         string `json:"status"`
	Deviation      string `json:"deviation"` // below | within | above | unknown
}

var numberRe = regexp.MustCompile(`-?\d+(?:[.,]\d+)?`)

// parseNum разбирает число с запятой или точкой
func parseNum(s string) (float64, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// classifyDeviation определяет положение значения относительно референса
func classifyDeviation(value, ref string) string {
	v, ok := parseNum(value)
	if !ok {
		return "unknown"
	}
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "unknown"
	}
	nums := numberRe.FindAllString(ref, -1)
	low := strings.ToLower(ref)

	// верхняя граница: «до», «менее», «не более», «<»
	if strings.Contains(low, "до ") || strings.Contains(low, "менее") ||
		strings.Contains(low, "не более") || strings.Contains(ref, "<") {
		if len(nums) >= 1 {
			if hi, ok := parseNum(nums[0]); ok {
				if v > hi {
					return "above"
				}
				return "within"
			}
		}
	}
	// нижняя граница: «более», «не менее», «>»
	if strings.Contains(low, "более") || strings.Contains(low, "не менее") ||
		strings.Contains(ref, ">") {
		if len(nums) >= 1 {
			if lo, ok := parseNum(nums[0]); ok {
				if v < lo {
					return "below"
				}
				return "within"
			}
		}
	}
	// диапазон из двух чисел
	if len(nums) >= 2 {
		lo, ok1 := parseNum(nums[0])
		hi, ok2 := parseNum(nums[1])
		if ok1 && ok2 {
			if lo > hi {
				lo, hi = hi, lo
			}
			switch {
			case v < lo:
				return "below"
			case v > hi:
				return "above"
			default:
				return "within"
			}
		}
	}
	return "unknown"
}

// --- контекст пациента ---

type patientContext struct {
	Profile      questionnaire
	Diagnoses    []map[string]any
	Medications  []map[string]any
	Measurements []map[string]any
	Indicators   []indicatorView
	HasAny       bool
}

func (p *patientContext) abnormal() []indicatorView {
	out := make([]indicatorView, 0)
	for _, ind := range p.Indicators {
		if ind.Deviation == "below" || ind.Deviation == "above" {
			out = append(out, ind)
		}
	}
	return out
}

// buildPatientContext собирает только подтверждённые данные пользователя
func (a *App) buildPatientContext(ctx context.Context, userID int) (*patientContext, error) {
	pc := &patientContext{
		Diagnoses:    []map[string]any{},
		Medications:  []map[string]any{},
		Measurements: []map[string]any{},
		Indicators:   []indicatorView{},
	}

	// анкета
	var birth *time.Time
	var weight *float64
	err := a.db.Pool.QueryRow(ctx,
		`SELECT birth_date, COALESCE(sex,''), COALESCE(known_diagnoses,''),
		        COALESCE(current_medications,''), COALESCE(cardiovascular_history,''),
		        weight_kg, COALESCE(risk_factors,'')
		 FROM health_profiles WHERE user_id=$1`, userID).
		Scan(&birth, &pc.Profile.Sex, &pc.Profile.KnownDiagnoses,
			&pc.Profile.CurrentMedications, &pc.Profile.CardiovascularHistory,
			&weight, &pc.Profile.RiskFactors)
	if err == nil {
		if birth != nil {
			pc.Profile.BirthDate = birth.Format("2006-01-02")
		}
		pc.Profile.WeightKg = weight
	}

	// диагнозы
	if rows, err := a.db.Pool.Query(ctx,
		`SELECT name, COALESCE(comment,'') FROM diagnoses
		 WHERE user_id=$1 ORDER BY created_at DESC`, userID); err == nil {
		for rows.Next() {
			var name, comment string
			if rows.Scan(&name, &comment) == nil {
				pc.Diagnoses = append(pc.Diagnoses, map[string]any{"name": name, "comment": comment})
			}
		}
		rows.Close()
	}

	// лекарства
	if rows, err := a.db.Pool.Query(ctx,
		`SELECT name, COALESCE(drug_class,''), COALESCE(frequency,''), status
		 FROM medications WHERE user_id=$1 ORDER BY created_at DESC`, userID); err == nil {
		for rows.Next() {
			var name, class, freq, status string
			if rows.Scan(&name, &class, &freq, &status) == nil {
				pc.Medications = append(pc.Medications, map[string]any{
					"name": name, "drug_class": class, "frequency": freq, "status": status,
				})
			}
		}
		rows.Close()
	}

	// измерения (последние 20)
	if rows, err := a.db.Pool.Query(ctx,
		`SELECT kind, measured_at, systolic, diastolic, pulse, weight
		 FROM measurements WHERE user_id=$1
		 ORDER BY measured_at DESC, id DESC LIMIT 20`, userID); err == nil {
		for rows.Next() {
			var kind string
			var measuredAt time.Time
			var systolic, diastolic, pulse *int
			var w *float64
			if rows.Scan(&kind, &measuredAt, &systolic, &diastolic, &pulse, &w) == nil {
				pc.Measurements = append(pc.Measurements, map[string]any{
					"kind": kind, "measured_at": measuredAt.Format("2006-01-02"),
					"systolic": intVal(systolic), "diastolic": intVal(diastolic),
					"pulse": intVal(pulse), "weight": fltVal(w),
				})
			}
		}
		rows.Close()
	}

	// подтверждённые показатели
	if rows, err := a.db.Pool.Query(ctx,
		`SELECT id, COALESCE(code,''), name, COALESCE(value,''), COALESCE(unit,''),
		        COALESCE(reference_range,''), test_date, source, verification_status
		 FROM lab_indicators
		 WHERE user_id=$1 AND verification_status IN ('confirmed','manual')
		 ORDER BY test_date DESC NULLS LAST, id DESC`, userID); err == nil {
		for rows.Next() {
			var iv indicatorView
			var testDate *time.Time
			if rows.Scan(&iv.ID, &iv.Code, &iv.Name, &iv.Value, &iv.Unit,
				&iv.ReferenceRange, &testDate, &iv.Source, &iv.Status) == nil {
				if testDate != nil {
					iv.TestDate = testDate.Format("2006-01-02")
				}
				iv.ReferenceRange = ai.ReferenceFor(iv.Code, iv.ReferenceRange)
				iv.Deviation = classifyDeviation(iv.Value, iv.ReferenceRange)
				pc.Indicators = append(pc.Indicators, iv)
			}
		}
		rows.Close()
	}

	pc.HasAny = len(pc.Indicators) > 0 || len(pc.Measurements) > 0 ||
		len(pc.Diagnoses) > 0 || len(pc.Medications) > 0 ||
		pc.Profile.Sex != "" || pc.Profile.BirthDate != "" ||
		pc.Profile.KnownDiagnoses != "" || pc.Profile.CurrentMedications != "" ||
		pc.Profile.RiskFactors != ""
	return pc, nil
}

// contextText формирует компактное текстовое представление для промпта
func (p *patientContext) contextText() string {
	var b strings.Builder

	b.WriteString("ПРОФИЛЬ:\n")
	writeIf := func(label, val string) {
		if strings.TrimSpace(val) != "" {
			fmt.Fprintf(&b, "- %s: %s\n", label, val)
		}
	}
	writeIf("Дата рождения", p.Profile.BirthDate)
	writeIf("Пол", p.Profile.Sex)
	writeIf("Семейный анамнез ССЗ", p.Profile.CardiovascularHistory)
	writeIf("Диагнозы из анкеты", p.Profile.KnownDiagnoses)
	writeIf("Лекарства из анкеты", p.Profile.CurrentMedications)
	writeIf("Факторы риска", p.Profile.RiskFactors)

	if len(p.Diagnoses) > 0 {
		b.WriteString("\nДИАГНОЗЫ:\n")
		for _, d := range p.Diagnoses {
			fmt.Fprintf(&b, "- %v", d["name"])
			if c, _ := d["comment"].(string); strings.TrimSpace(c) != "" {
				fmt.Fprintf(&b, " (%s)", c)
			}
			b.WriteString("\n")
		}
	}

	if len(p.Medications) > 0 {
		b.WriteString("\nЛЕКАРСТВА:\n")
		for _, m := range p.Medications {
			fmt.Fprintf(&b, "- %v", m["name"])
			if c, _ := m["drug_class"].(string); strings.TrimSpace(c) != "" {
				fmt.Fprintf(&b, ", класс: %s", c)
			}
			if f, _ := m["frequency"].(string); strings.TrimSpace(f) != "" {
				fmt.Fprintf(&b, ", частота: %s", f)
			}
			if s, _ := m["status"].(string); s == "taking" {
				b.WriteString(", принимает")
			}
			b.WriteString("\n")
		}
	}

	if len(p.Measurements) > 0 {
		b.WriteString("\nИЗМЕРЕНИЯ (последние):\n")
		for _, m := range p.Measurements {
			fmt.Fprintf(&b, "- %v: %s\n", m["measured_at"], formatMeasurement(m))
		}
	}

	if len(p.Indicators) > 0 {
		b.WriteString("\nПОДТВЕРЖДЁННЫЕ ПОКАЗАТЕЛИ:\n")
		for _, ind := range p.Indicators {
			fmt.Fprintf(&b, "- %s: %s %s", ind.Name, ind.Value, ind.Unit)
			if ind.ReferenceRange != "" {
				fmt.Fprintf(&b, " (реф. %s)", ind.ReferenceRange)
			}
			if ind.TestDate != "" {
				fmt.Fprintf(&b, ", %s", ind.TestDate)
			}
			fmt.Fprintf(&b, " — %s\n", deviationLabel(ind.Deviation))
		}
	}

	if b.Len() == 0 {
		return "Данных в сводке пока нет."
	}
	return b.String()
}

func formatMeasurement(m map[string]any) string {
	kind, _ := m["kind"].(string)
	switch kind {
	case "pressure":
		return fmt.Sprintf("давление %v/%v", deref(m["systolic"]), deref(m["diastolic"]))
	case "pulse":
		return fmt.Sprintf("пульс %v", deref(m["pulse"]))
	case "weight":
		return fmt.Sprintf("вес %v кг", deref(m["weight"]))
	}
	return kind
}

func deref(v any) any {
	if v == nil {
		return "—"
	}
	return v
}

// intVal разыменовывает *int в значение или nil
func intVal(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

// fltVal разыменовывает *float64 в значение или nil
func fltVal(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func deviationLabel(d string) string {
	switch d {
	case "below":
		return "ниже референса"
	case "above":
		return "выше референса"
	case "within":
		return "в пределах референса"
	}
	return "референс не указан"
}

// --- эндпоинты ---

// getSummary — детерминированная личная сводка (FR10, FR8), без ИИ
func (a *App) getSummary(w http.ResponseWriter, r *http.Request) {
	pc, err := a.buildPatientContext(r.Context(), userIDFrom(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось собрать сводку")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"profile":      pc.Profile,
		"diagnoses":    pc.Diagnoses,
		"medications":  pc.Medications,
		"measurements": pc.Measurements,
		"indicators":   pc.Indicators,
		"abnormal":     pc.abnormal(),
		"empty":        !pc.HasAny,
		"updated_at":   time.Now().Format(time.RFC3339),
	})
}

// postOverview — короткий AI-обзор сводки и подготовка к визиту (FR8, FR15)
func (a *App) postOverview(w http.ResponseWriter, r *http.Request) {
	pc, err := a.buildPatientContext(r.Context(), userIDFrom(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось собрать сводку")
		return
	}
	if !pc.HasAny {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"text":         "Пока недостаточно данных для обзора. Загрузите анализы или заполните анкету и измерения.",
			"generated_at": time.Now().Format(time.RFC3339),
			"empty":        true,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	answer, err := a.ai.Chat(ctx, ai.OverviewSystemPrompt, pc.contextText(), nil)
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, "AI недоступен: "+err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"text":         ai.CheckOverviewAnswer(answer),
		"generated_at": time.Now().Format(time.RFC3339),
		"empty":        false,
	})
}

type chatRequest struct {
	Question string           `json:"question"`
	History  []ai.ChatMessage `json:"history"`
}

// postChat — AI-чат по личной сводке (FR15)
func (a *App) postChat(w http.ResponseWriter, r *http.Request) {
	var in chatRequest
	if !decode(w, r, &in) {
		return
	}
	in.Question = strings.TrimSpace(in.Question)
	if in.Question == "" {
		httpx.Error(w, http.StatusBadRequest, "Пустой вопрос")
		return
	}
	if len(in.History) > 10 {
		in.History = in.History[len(in.History)-10:]
	}

	pc, err := a.buildPatientContext(r.Context(), userIDFrom(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось собрать сводку")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	question := pc.contextText() + "\n\nВОПРОС ПОЛЬЗОВАТЕЛЯ:\n" + in.Question
	answer, err := a.ai.Chat(ctx, ai.ChatSystemPrompt, question, in.History)
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, "AI недоступен: "+err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"answer": ai.CheckChatAnswer(answer),
	})
}
