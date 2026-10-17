package app

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-med/backend/internal/ai"
	"ai-med/backend/internal/docparse"
	"ai-med/backend/internal/httpx"
)

const maxUploadBytes = 20 << 20 // 20 МБ

type documentJSON struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Mime       string    `json:"mime"`
	Status     string    `json:"status"`
	LabName    string    `json:"lab_name"`
	TestDate   string    `json:"test_date"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (a *App) scanDocument(sc interface {
	Scan(...any) error
}) (documentJSON, error) {
	var d documentJSON
	var labName *string
	var testDate *time.Time
	err := sc.Scan(&d.ID, &d.Filename, &d.Mime, &d.Status, &labName, &testDate, &d.UploadedAt)
	if labName != nil {
		d.LabName = *labName
	}
	if testDate != nil {
		d.TestDate = testDate.Format("2006-01-02")
	}
	return d, err
}

const documentColumns = `id, filename, COALESCE(mime,''), status, lab_name, test_date, uploaded_at`

func (a *App) listDocuments(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	rows, err := a.db.Pool.Query(r.Context(),
		`SELECT `+documentColumns+`, stored_path FROM lab_documents
		 WHERE user_id=$1 ORDER BY uploaded_at DESC`, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить документы")
		return
	}
	defer rows.Close()

	items := make([]documentJSON, 0)
	for rows.Next() {
		var d documentJSON
		var labName *string
		var testDate *time.Time
		var storedPath string
		if err := rows.Scan(&d.ID, &d.Filename, &d.Mime, &d.Status, &labName, &testDate, &d.UploadedAt, &storedPath); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Ошибка чтения данных")
			return
		}
		if labName != nil {
			d.LabName = *labName
		}
		if testDate != nil {
			d.TestDate = testDate.Format("2006-01-02")
		}
		items = append(items, d)
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (a *App) getDocument(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	row := a.db.Pool.QueryRow(r.Context(),
		`SELECT `+documentColumns+` FROM lab_documents WHERE id=$1 AND user_id=$2`, did, id)
	d, err := a.scanDocument(row)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Документ не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, d)
}

func (a *App) getDocumentStatus(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var status string
	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT status FROM lab_documents WHERE id=$1 AND user_id=$2`, did, id).Scan(&status)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Документ не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": status})
}

func (a *App) getDocumentFile(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var storedPath, mime, filename string
	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT stored_path, COALESCE(mime,'application/octet-stream'), filename
		 FROM lab_documents WHERE id=$1 AND user_id=$2`, did, id).
		Scan(&storedPath, &mime, &filename)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Документ не найден")
		return
	}
	rc, err := a.store.Open(storedPath)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Файл недоступен")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func (a *App) deleteDocument(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var storedPath string
	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT stored_path FROM lab_documents WHERE id=$1 AND user_id=$2`, did, id).Scan(&storedPath)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Документ не найден")
		return
	}
	if _, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM lab_indicators WHERE document_id=$1 AND user_id=$2`, did, id); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить показатели")
		return
	}
	if _, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM lab_documents WHERE id=$1 AND user_id=$2`, did, id); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить документ")
		return
	}
	_ = a.store.Delete(storedPath)
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// createDocument — загрузка FR5 и извлечение FR6
func (a *App) createDocument(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Файл слишком большой или повреждён (макс. 20 МБ)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Файл не передан")
		return
	}
	defer file.Close()

	format, ok := docparse.Detect(header.Filename)
	if !ok {
		httpx.Error(w, http.StatusBadRequest,
			"Неподдерживаемый формат. Разрешены: "+strings.Join(docparse.SupportedExtensions(), ", "))
		return
	}

	storedPath, err := a.store.Save(id, header.Filename, file)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось сохранить файл")
		return
	}

	var docID int
	if err := a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO lab_documents (user_id, filename, stored_path, mime, status)
		 VALUES ($1,$2,$3,$4,'processing') RETURNING id`,
		id, header.Filename, storedPath, format.MIME).Scan(&docID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось создать документ")
		return
	}

	// читаем сохранённый файл для анализа
	rc, err := a.store.Open(storedPath)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось прочитать файл")
		return
	}
	raw, _ := io.ReadAll(rc)
	_ = rc.Close()

	var content ai.Content
	if format.Text {
		text, err := docparse.DocxTextBytes(raw)
		if err != nil {
			a.markFailed(r.Context(), docID, id)
			httpx.JSON(w, http.StatusOK, map[string]any{
				"document":   a.mustGetDocument(r.Context(), id, docID),
				"indicators": []any{},
				"warning":    "Не удалось прочитать документ",
			})
			return
		}
		content = ai.Content{Text: text, Filename: header.Filename}
	} else if format.MIME == "application/pdf" {
		// OpenAI-совместимые провайдеры не принимают PDF напрямую —
		// растеризуем страницы в PNG и отправляем как изображения
		pages, perr := docparse.RasterizePDF(raw)
		if perr != nil {
			a.markFailed(r.Context(), docID, id)
			httpx.JSON(w, http.StatusOK, map[string]any{
				"document":   a.mustGetDocument(r.Context(), id, docID),
				"indicators": []any{},
				"warning":    "Не удалось подготовить PDF: " + perr.Error(),
			})
			return
		}
		imgs := make([]ai.ImagePart, 0, len(pages))
		for _, p := range pages {
			imgs = append(imgs, ai.ImagePart{Data: p.Data, MIME: p.MIME})
		}
		content = ai.Content{Images: imgs, Filename: header.Filename}
	} else {
		content = ai.Content{FileBytes: raw, MIME: format.MIME, Filename: header.Filename}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, extractErr := a.ai.Extract(ctx, content)
	if extractErr != nil {
		a.markFailed(r.Context(), docID, id)
		httpx.JSON(w, http.StatusOK, map[string]any{
			"document":   a.mustGetDocument(r.Context(), id, docID),
			"indicators": []any{},
			"warning":    "Не удалось распознать документ: " + extractErr.Error(),
		})
		return
	}

	result = ai.ApplyExtractionGuardrails(result)

	docTestDate := parseDateOrNil(result.TestDate)
	for _, ind := range result.Indicators {
		testDate := parseDateOrNil(ind.TestDate)
		if testDate == nil {
			testDate = docTestDate
		}
		_, _ = a.db.Pool.Exec(r.Context(),
			`INSERT INTO lab_indicators
			   (user_id, document_id, code, name, value, unit, reference_range,
			    test_date, source, verification_status, confidence)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'ocr','needs_review',$9)`,
			id, docID, ind.Code, ind.Name, ind.Value, ind.Unit,
			ai.ReferenceFor(ind.Code, ind.ReferenceRange),
			testDate, ind.Confidence)
	}

	status := "failed"
	if len(result.Indicators) > 0 {
		status = "needs_review"
	}
	_, _ = a.db.Pool.Exec(r.Context(),
		`UPDATE lab_documents SET status=$1, lab_name=$2, test_date=$3 WHERE id=$4`,
		status, nullableStr(result.LabName), docTestDate, docID)

	indicators := a.listIndicatorsForDoc(r.Context(), id, docID)
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"document":   a.mustGetDocument(r.Context(), id, docID),
		"indicators": indicators,
	})
}

func nullableStr(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func parseDateOrNil(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	// формат дд.мм.гггг
	if t, err := time.Parse("02.01.2006", s); err == nil {
		return t
	}
	return nil
}

func (a *App) markFailed(ctx context.Context, docID, userID int) {
	_, _ = a.db.Pool.Exec(ctx,
		`UPDATE lab_documents SET status='failed' WHERE id=$1 AND user_id=$2`, docID, userID)
}

func (a *App) mustGetDocument(ctx context.Context, userID, docID int) documentJSON {
	row := a.db.Pool.QueryRow(ctx,
		`SELECT `+documentColumns+` FROM lab_documents WHERE id=$1 AND user_id=$2`, docID, userID)
	d, _ := a.scanDocument(row)
	return d
}

// --- FR6/FR7: показатели ---

func (a *App) listIndicatorsForDoc(ctx context.Context, userID, docID int) []map[string]any {
	rows, err := a.db.Pool.Query(ctx,
		`SELECT id, COALESCE(code,''), name, COALESCE(value,''), COALESCE(unit,''),
		        COALESCE(reference_range,''), test_date, source, verification_status,
		        confidence, updated_at
		 FROM lab_indicators WHERE user_id=$1 AND document_id=$2
		 ORDER BY id ASC`, userID, docID)
	if err != nil {
		return []map[string]any{}
	}
	defer rows.Close()
	return scanIndicatorRows(rows)
}

func scanIndicatorRows(rows interface {
	Next() bool
	Scan(...any) error
}) []map[string]any {
	items := make([]map[string]any, 0)
	for rows.Next() {
		var iid int
		var code, name, value, unit, ref, source, status string
		var testDate *time.Time
		var confidence *float64
		var updated time.Time
		if err := rows.Scan(&iid, &code, &name, &value, &unit, &ref, &testDate,
			&source, &status, &confidence, &updated); err != nil {
			continue
		}
		var dateStr any
		if testDate != nil {
			dateStr = testDate.Format("2006-01-02")
		}
		items = append(items, map[string]any{
			"id": iid, "code": code, "name": name, "value": value, "unit": unit,
			"reference_range": ai.ReferenceFor(code, ref), "test_date": dateStr, "source": source,
			"verification_status": status, "confidence": confidence, "updated_at": updated,
		})
	}
	return items
}

func (a *App) listDocumentIndicators(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	httpx.JSON(w, http.StatusOK, a.listIndicatorsForDoc(r.Context(), id, did))
}

func (a *App) confirmAllIndicators(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	did, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE lab_indicators SET verification_status='confirmed', updated_at=CURRENT_TIMESTAMP
		 WHERE document_id=$1 AND user_id=$2`, did, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось подтвердить показатели")
		return
	}
	_, _ = a.db.Pool.Exec(r.Context(),
		`UPDATE lab_documents SET status='saved' WHERE id=$1 AND user_id=$2`, did, id)
	httpx.JSON(w, http.StatusOK, map[string]any{"confirmed": tag.RowsAffected()})
}

type indicatorIn struct {
	Name           *string `json:"name"`
	Value          *string `json:"value"`
	Unit           *string `json:"unit"`
	ReferenceRange *string `json:"reference_range"`
	TestDate       *string `json:"test_date"`
}

func (a *App) updateIndicator(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	iid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var in indicatorIn
	if !decode(w, r, &in) {
		return
	}
	var name, value, unit, ref any = nil, nil, nil, nil
	if in.Name != nil {
		name = *in.Name
	}
	if in.Value != nil {
		value = *in.Value
	}
	if in.Unit != nil {
		unit = *in.Unit
	}
	if in.ReferenceRange != nil {
		ref = *in.ReferenceRange
	}
	var testDate any
	if in.TestDate != nil {
		testDate = parseDateOrNil(*in.TestDate)
	}

	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE lab_indicators SET
		    name = COALESCE($1, name),
		    value = COALESCE($2, value),
		    unit = COALESCE($3, unit),
		    reference_range = COALESCE($4, reference_range),
		    test_date = COALESCE($5, test_date),
		    updated_at = CURRENT_TIMESTAMP
		 WHERE id=$6 AND user_id=$7`,
		name, value, unit, ref, testDate, iid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось обновить показатель")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Показатель не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) deleteIndicator(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	iid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`DELETE FROM lab_indicators WHERE id=$1 AND user_id=$2`, iid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось удалить показатель")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Показатель не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) setIndicatorStatus(w http.ResponseWriter, r *http.Request, status string) {
	id := userIDFrom(r)
	iid, ok := urlID(r, "id")
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	tag, err := a.db.Pool.Exec(r.Context(),
		`UPDATE lab_indicators SET verification_status=$1, updated_at=CURRENT_TIMESTAMP
		 WHERE id=$2 AND user_id=$3`, status, iid, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось изменить статус")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, "Показатель не найден")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"verification_status": status})
}

func (a *App) confirmIndicator(w http.ResponseWriter, r *http.Request) {
	a.setIndicatorStatus(w, r, "confirmed")
}

func (a *App) needsReviewIndicator(w http.ResponseWriter, r *http.Request) {
	a.setIndicatorStatus(w, r, "needs_review")
}
