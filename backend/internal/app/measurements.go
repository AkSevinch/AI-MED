package app

import (
	"net/http"
	"strings"
	"time"

	"ai-med/backend/internal/httpx"
)

// --- FR4: домашние измерения ---

type measurementIn struct {
	Kind       string   `json:"kind"`
	MeasuredAt string   `json:"measured_at"`
	Systolic   *int     `json:"systolic"`
	Diastolic  *int     `json:"diastolic"`
	Pulse      *int     `json:"pulse"`
	Weight     *float64 `json:"weight"`
	Comment    string   `json:"comment"`
}

func validKind(k string) bool {
	switch strings.ToLower(strings.TrimSpace(k)) {
	case "pressure", "pulse", "weight":
		return true
	}
	return false
}

// validateMeasurement проверяет правила SRS FR4
func validateMeasurement(in *measurementIn) string {
	in.Kind = strings.ToLower(strings.TrimSpace(in.Kind))
	if !validKind(in.Kind) {
		return "Недопустимый тип измерения"
	}
	if strings.TrimSpace(in.MeasuredAt) == "" {
		return "Дата измерения обязательна"
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(in.MeasuredAt)); err != nil {
		return "Некорректная дата измерения"
	}
	switch in.Kind {
	case "pressure":
		if in.Systolic == nil || in.Diastolic == nil {
			return "Для давления нужны систолическое и диастолическое значения"
		}
		if *in.Systolic <= 0 || *in.Diastolic <= 0 {
			return "Значения давления должны быть положительными числами"
		}
	case "pulse":
		if in.Pulse == nil || *in.Pulse <= 0 {
			return "Пульс должен быть положительным числом"
		}
	case "weight":
		if in.Weight == nil || *in.Weight <= 0 {
			return "Вес должен быть положительным числом"
		}
	}
	if in.Pulse != nil && *in.Pulse <= 0 {
		return "Пульс должен быть положительным числом"
	}
	return ""
}

func (a *App) listMeasurements(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	rows, err := a.db.Pool.Query(r.Context(),
		`SELECT id, kind, measured_at, systolic, diastolic, pulse, weight,
		        COALESCE(comment,''), created_at
		 FROM measurements WHERE user_id=$1
		 ORDER BY measured_at ASC, id ASC`, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить измерения")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var mid int
		var kind, comment string
		var measuredAt time.Time
		var created time.Time
		var systolic, diastolic, pulse *int
		var weight *float64
		if err := rows.Scan(&mid, &kind, &measuredAt, &systolic, &diastolic, &pulse,
			&weight, &comment, &created); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Ошибка чтения данных")
			return
		}
		items = append(items, map[string]any{
			"id": mid, "kind": kind, "measured_at": measuredAt.Format("2006-01-02"),
			"systolic": systolic, "diastolic": diastolic, "pulse": pulse,
			"weight": weight, "comment": comment, "created_at": created,
		})
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (a *App) createMeasurement(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var in measurementIn
	if !decode(w, r, &in) {
		return
	}
	if msg := validateMeasurement(&in); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	var newID int
	if err := a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO measurements (user_id, kind, measured_at, systolic, diastolic, pulse, weight, comment)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		id, in.Kind, strings.TrimSpace(in.MeasuredAt), in.Systolic, in.Diastolic,
		in.Pulse, in.Weight, in.Comment).Scan(&newID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось сохранить измерение")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": newID, "kind": in.Kind})
}

func (a *App) updateMeasurement(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	mid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var in measurementIn
	if !decode(w, r, &in) {
		return
	}
	if msg := validateMeasurement(&in); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE measurements SET kind=$1, measured_at=$2, systolic=$3, diastolic=$4,
		    pulse=$5, weight=$6, comment=$7
		 WHERE id=$8 AND user_id=$9`,
		in.Kind, strings.TrimSpace(in.MeasuredAt), in.Systolic, in.Diastolic,
		in.Pulse, in.Weight, in.Comment, mid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось обновить измерение")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Измерение не найдено")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"id": mid, "kind": in.Kind})
}

func (a *App) deleteMeasurement(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	mid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM measurements WHERE id=$1 AND user_id=$2`, mid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить измерение")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Измерение не найдено")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
