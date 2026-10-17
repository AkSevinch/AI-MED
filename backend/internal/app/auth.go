package app

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ai-med/backend/internal/httpx"
)

// ConsentVersion — текущая версия текста согласия
const ConsentVersion = "1.0"

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validEmail(e string) bool {
	_, err := mail.ParseAddress(e)
	return err == nil
}

func (a *App) setSession(w http.ResponseWriter, r *http.Request, userID int) error {
	s := a.session(r)
	s.Values["user_id"] = userID
	return s.Save(r, w)
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if !validEmail(in.Email) {
		httpx.Error(w, http.StatusBadRequest, "Некорректный email")
		return
	}
	if len(in.Password) < 8 {
		httpx.Error(w, http.StatusBadRequest, "Пароль должен быть не короче 8 символов")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Ошибка обработки пароля")
		return
	}

	var id int
	err = a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id`, in.Email, string(hash)).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			httpx.Error(w, http.StatusConflict, "Пользователь с таким email уже существует")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	if err := a.setSession(w, r, id); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось создать сессию")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"id":          id,
		"email":       in.Email,
		"has_consent": false,
	})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	var id int
	var hash string
	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT id, password_hash FROM users WHERE email=$1`, in.Email).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		httpx.Error(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	if err := a.setSession(w, r, id); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось создать сессию")
		return
	}

	payload, err := a.userPayload(r, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить пользователя")
		return
	}
	httpx.JSON(w, http.StatusOK, payload)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	s := a.session(r)
	s.Options.MaxAge = -1
	_ = s.Save(r, w)
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	if id == 0 {
		httpx.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	payload, err := a.userPayload(r, id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось загрузить пользователя")
		return
	}
	httpx.JSON(w, http.StatusOK, payload)
}

func (a *App) userPayload(r *http.Request, id int) (map[string]any, error) {
	var email string
	var createdAt time.Time
	if err := a.db.Pool.QueryRow(r.Context(),
		`SELECT email, created_at FROM users WHERE id=$1`, id).Scan(&email, &createdAt); err != nil {
		return nil, err
	}
	var hasConsent bool
	_ = a.db.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM consents WHERE user_id=$1)`, id).Scan(&hasConsent)

	return map[string]any{
		"authenticated": true,
		"id":            id,
		"email":         email,
		"has_consent":   hasConsent,
		"created_at":    createdAt,
	}, nil
}

func (a *App) getConsent(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var version string
	var acceptedAt time.Time
	err := a.db.Pool.QueryRow(r.Context(),
		`SELECT version, accepted_at FROM consents WHERE user_id=$1
		 ORDER BY accepted_at DESC LIMIT 1`, id).Scan(&version, &acceptedAt)
	if err != nil {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"accepted":   false,
			"version":    ConsentVersion,
			"is_current": false,
		})
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"accepted":    true,
		"version":     version,
		"accepted_at": acceptedAt,
		"is_current":  version == ConsentVersion,
	})
}

func (a *App) postConsent(w http.ResponseWriter, r *http.Request) {
	id := userIDFrom(r)
	var in struct {
		Version string `json:"version"`
	}
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Version) == "" {
		in.Version = ConsentVersion
	}
	var acceptedAt time.Time
	if err := a.db.Pool.QueryRow(r.Context(),
		`INSERT INTO consents (user_id, version) VALUES ($1, $2)
		 RETURNING accepted_at`, id, in.Version).Scan(&acceptedAt); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Не удалось сохранить согласие")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"accepted":    true,
		"version":     in.Version,
		"accepted_at": acceptedAt,
	})
}
