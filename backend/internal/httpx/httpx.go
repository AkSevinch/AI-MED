package httpx

import (
	"encoding/json"
	"io"
	"net/http"
)

// JSON пишет v как JSON-ответ с заданным кодом статуса
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// Error пишет JSON-конверт с ошибкой
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Decode читает JSON тела запроса в v, отклоняя неизвестные поля
// и ограничивая размер тела
func Decode(w http.ResponseWriter, r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
