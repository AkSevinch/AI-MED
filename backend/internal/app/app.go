package app

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"

	"ai-med/backend/internal/ai"
	"ai-med/backend/internal/config"
	"ai-med/backend/internal/db"
	"ai-med/backend/internal/httpx"
	"ai-med/backend/internal/storage"
)

const sessionName = "aimed_session"

// App связывает конфигурацию, хранилище, БД и провайдер ИИ с HTTP-роутером
type App struct {
	cfg      *config.Config
	db       *db.DB
	store    storage.Storage
	ai       ai.Provider
	sessions *sessions.CookieStore
}

// New создаёт App
func New(cfg *config.Config, database *db.DB, store storage.Storage, provider ai.Provider) *App {
	cs := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	cs.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	}
	return &App{cfg: cfg, db: database, store: store, ai: provider, sessions: cs}
}

// Router собирает chi-роутер со всеми маршрутами и middleware
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(a.cors)
	r.Use(a.loadSession)

	r.Get("/api/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"status":   "ok",
			"provider": a.ai.Name(),
		})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", a.register)
		r.Post("/login", a.login)
		r.Post("/logout", a.logout)
		r.Get("/me", a.me)
	})

	r.Group(func(r chi.Router) {
		r.Use(a.requireAuth)

		r.Get("/api/consent", a.getConsent)
		r.Post("/api/consent", a.postConsent)

		r.Group(func(r chi.Router) {
			r.Use(a.requireConsent)

			r.Get("/api/profile/questionnaire", a.getQuestionnaire)
			r.Put("/api/profile/questionnaire", a.putQuestionnaire)

			r.Get("/api/diagnoses", a.listDiagnoses)
			r.Post("/api/diagnoses", a.createDiagnosis)
			r.Put("/api/diagnoses/{id}", a.updateDiagnosis)
			r.Delete("/api/diagnoses/{id}", a.deleteDiagnosis)

			r.Get("/api/medications", a.listMedications)
			r.Post("/api/medications", a.createMedication)
			r.Put("/api/medications/{id}", a.updateMedication)
			r.Delete("/api/medications/{id}", a.deleteMedication)

			r.Get("/api/measurements", a.listMeasurements)
			r.Post("/api/measurements", a.createMeasurement)
			r.Put("/api/measurements/{id}", a.updateMeasurement)
			r.Delete("/api/measurements/{id}", a.deleteMeasurement)

			r.Get("/api/documents", a.listDocuments)
			r.Post("/api/documents", a.createDocument)
			r.Get("/api/documents/{id}", a.getDocument)
			r.Delete("/api/documents/{id}", a.deleteDocument)
			r.Get("/api/documents/{id}/file", a.getDocumentFile)
			r.Get("/api/documents/{id}/status", a.getDocumentStatus)
			r.Get("/api/documents/{id}/indicators", a.listDocumentIndicators)
			r.Post("/api/documents/{id}/confirm-all", a.confirmAllIndicators)

			r.Put("/api/indicators/{id}", a.updateIndicator)
			r.Delete("/api/indicators/{id}", a.deleteIndicator)
			r.Post("/api/indicators/{id}/confirm", a.confirmIndicator)
			r.Post("/api/indicators/{id}/needs-review", a.needsReviewIndicator)

			r.Get("/api/summary", a.getSummary)
			r.Post("/api/summary/overview", a.postOverview)
			r.Post("/api/chat", a.postChat)
		})
	})

	return r
}

// --- посредники ---

type ctxKey int

const (
	userIDKey ctxKey = iota
)

func userIDFrom(r *http.Request) int {
	if v, ok := r.Context().Value(userIDKey).(int); ok {
		return v
	}
	return 0
}

func (a *App) session(r *http.Request) *sessions.Session {
	s, _ := a.sessions.Get(r, sessionName)
	return s
}

func (a *App) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := 0
		if s := a.session(r); s != nil {
			if v, ok := s.Values["user_id"].(int); ok {
				id = v
			}
		}
		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userIDFrom(r) == 0 {
			httpx.Error(w, http.StatusUnauthorized, "Требуется вход в систему")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) requireConsent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := userIDFrom(r)
		var exists bool
		err := a.db.Pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM consents WHERE user_id=$1)`, id).Scan(&exists)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Не удалось проверить согласие")
			return
		}
		if !exists {
			httpx.Error(w, http.StatusForbidden, "Необходимо подтвердить согласие на обработку данных")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) cors(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(a.cfg.CORSOrigins))
	for _, o := range a.cfg.CORSOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- вспомогательные функции ---

// urlID парсит параметр URL chi в int
func urlID(r *http.Request, key string) (int, bool) {
	raw := strings.TrimSpace(chi.URLParam(r, key))
	if raw == "" {
		return 0, false
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// decode пишет 400 при ошибке разбора тела
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := httpx.Decode(w, r, v); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Некорректное тело запроса: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	httpx.JSON(w, http.StatusOK, v)
}
