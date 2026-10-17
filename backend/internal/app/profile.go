package app

import (
	"net/http"
	"strings"
	"time"

	"ai-med/backend/internal/httpx"
)

func dateArg(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

// --- FR2: анкета состояния здоровья ---

type questionnaire struct {
	BirthDate             string   `json:"birth_date"`
	Sex                   string   `json:"sex"`
	KnownDiagnoses        string   `json:"known_diagnoses"`
	CurrentMedications    string   `json:"current_medications"`
	CardiovascularHistory string   `json:"cardiovascular_history"`
	WeightKg              *float64 `json:"weight_kg"`
	RiskFactors           string   `json:"risk_factors"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

func (a *App) getQuestionnaire(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var q questionnaire
	var birth *time.Time
	var weight *float64
	var updated time.Time

	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT birth_date, sex, known_diagnoses, current_medications,
		        cardiovascular_history, weight_kg, risk_factors, updated_at
		 FROM health_profiles WHERE user_id=$1`, id).
		Scan(&birth, &q.Sex, &q.KnownDiagnoses, &q.CurrentMedications,
			&q.CardiovascularHistory, &weight, &q.RiskFactors, &updated)
	if err == nil {
		if birth != nil {
			q.BirthDate = birth.Format("2006-01-02")
		}
		q.WeightKg = weight
		q.UpdatedAt = updated.Format(time.RFC3339)
	}
	httpx.JSON(w, http.StatusOK, q)
}

func (a *App) putQuestionnaire(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var in questionnaire
	if !decode(w, r, &in) {
		return
	}
	_, err := a.db.Pool.Exec(r.Context(),
		`INSERT INTO health_profiles
		   (user_id, birth_date, sex, known_diagnoses, current_medications,
		    cardiovascular_history, weight_kg, risk_factors, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id) DO UPDATE SET
		   birth_date=EXCLUDED.birth_date,
		   sex=EXCLUDED.sex,
		   known_diagnoses=EXCLUDED.known_diagnoses,
		   current_medications=EXCLUDED.current_medications,
		   cardiovascular_history=EXCLUDED.cardiovascular_history,
		   weight_kg=EXCLUDED.weight_kg,
		   risk_factors=EXCLUDED.risk_factors,
		   updated_at=CURRENT_TIMESTAMP`,
		id, dateArg(in.BirthDate), in.Sex, in.KnownDiagnoses, in.CurrentMedications,
		in.CardiovascularHistory, in.WeightKg, in.RiskFactors)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось сохранить анкету")
		return
	}
	a.getQuestionnaire(w, r)
}

// --- FR3: диагнозы ---

type diagnosisIn struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

func (a *App) listDiagnoses(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	rows, err := a.db.Pool.Query(r.Context(),
		`SELECT id, name, COALESCE(comment,''), created_at, updated_at
		 FROM diagnoses WHERE user_id=$1 ORDER BY created_at DESC`, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить диагнозы")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var did int
		var name, comment string
		var created, updated time.Time
		if err := rows.Scan(&did, &name, &comment, &created, &updated); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Ошибка чтения данных")
			return
		}
		items = append(items, map[string]any{
			"id": did, "name": name, "comment": comment,
			"created_at": created, "updated_at": updated,
		})
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (a *App) createDiagnosis(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var in diagnosisIn
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		httpx.Error(w, http.StatusBadRequest, "Название диагноза обязательно")
		return
	}
	var newID int
	if err := a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO diagnoses (user_id, name, comment) VALUES ($1,$2,$3) RETURNING id`,
		id, in.Name, in.Comment).Scan(&newID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось добавить диагноз")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": newID, "name": in.Name, "comment": in.Comment})
}

func (a *App) updateDiagnosis(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var in diagnosisIn
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		httpx.Error(w, http.StatusBadRequest, "Название диагноза обязательно")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE diagnoses SET name=$1, comment=$2, updated_at=CURRENT_TIMESTAMP
		 WHERE id=$3 AND user_id=$4`, in.Name, in.Comment, did, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось обновить диагноз")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Диагноз не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"id": did, "name": in.Name, "comment": in.Comment})
}

func (a *App) deleteDiagnosis(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM diagnoses WHERE id=$1 AND user_id=$2`, did, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить диагноз")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Диагноз не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- FR3: лекарства ---

type medicationIn struct {
	Name      string `json:"name"`
	DrugClass string `json:"drug_class"`
	Frequency string `json:"frequency"`
	Status    string `json:"status"`
	Comment   string `json:"comment"`
}

var medicationStatuses = map[string]bool{"taking": true, "paused": true, "unknown": true}

func normalizeMedStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if !medicationStatuses[s] {
		return "unknown"
	}
	return s
}

func (a *App) listMedications(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	rows, err := a.db.Pool.Query(r.Context(),
		`SELECT id, name, COALESCE(drug_class,''), COALESCE(frequency,''),
		        status, COALESCE(comment,''), created_at, updated_at
		 FROM medications WHERE user_id=$1 ORDER BY created_at DESC`, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить лекарства")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var mid int
		var name, drugClass, frequency, status, comment string
		var created, updated time.Time
		if err := rows.Scan(&mid, &name, &drugClass, &frequency, &status, &comment, &created, &updated); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Ошибка чтения данных")
			return
		}
		items = append(items, map[string]any{
			"id": mid, "name": name, "drug_class": drugClass, "frequency": frequency,
			"status": status, "comment": comment, "created_at": created, "updated_at": updated,
		})
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (a *App) createMedication(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var in medicationIn
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		httpx.Error(w, http.StatusBadRequest, "Название лекарства обязательно")
		return
	}
	status := normalizeMedStatus(in.Status)
	var newID int
	if err := a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO medications (user_id, name, drug_class, frequency, status, comment)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		id, in.Name, in.DrugClass, in.Frequency, status, in.Comment).Scan(&newID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось добавить лекарство")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": newID, "name": in.Name, "status": status})
}

func (a *App) updateMedication(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	mid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var in medicationIn
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		httpx.Error(w, http.StatusBadRequest, "Название лекарства обязательно")
		return
	}
	status := normalizeMedStatus(in.Status)
	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE medications SET name=$1, drug_class=$2, frequency=$3, status=$4,
		    comment=$5, updated_at=CURRENT_TIMESTAMP
		 WHERE id=$6 AND user_id=$7`,
		in.Name, in.DrugClass, in.Frequency, status, in.Comment, mid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось обновить лекарство")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Лекарство не найдено")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"id": mid, "name": in.Name, "status": status})
}

func (a *App) deleteMedication(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	mid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM medications WHERE id=$1 AND user_id=$2`, mid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить лекарство")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Лекарство не найдено")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
