# 🏥 AI MED — единая сводка здоровья

> **AI MED помогает собрать разрозненные медицинские данные в одну понятную сводку. Сервис не ставит диагнозы, не назначает лечение и не анализирует снимки.**

[![Next.js](https://img.shields.io/badge/Next.js-15.4.5-black)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0-blue)](https://www.typescriptlang.org/)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791)](https://www.postgresql.org/)

---

## 🎯 Проблема

У человека накапливаются результаты анализов, измерения и история приёма
лекарств из разных источников. Собрать это в одну картину и понять, что уже
подтверждено, а что требует уточнения, — сложно. Онлайн-сервисы часто вместо
структурирования данных пытаются «интерпретировать» их, что небезопасно.

---

## 💡 Решение

**AI MED** — это личный кабинет, который:

- принимает бланки анализов (PDF, JPG, PNG, DOCX);
- распознаёт числовые показатели с помощью ИИ-**vision**;
- **не включает ни одно значение в сводку без ручного подтверждения** пользователем;
- хранит анкету, диагнозы, лекарства и домашние измерения как контекст.

### Что сервис делает

- 📄 Загрузка и распознавание бланков анализов
- ✅ Ручная проверка каждого распознанного показателя
- 👤 Анкета здоровья (пол, дата рождения, анамнез, факторы риска)
- 🩺 Ведение списка установленных врачом диагнозов
- 💊 Ведение списка принимаемых лекарств
- 📊 Домашние измерения: давление, пульс, вес
- 🧠 ИИ-обзор подтверждённых данных — короткая карточка на «Обзоре»
- 💬 ИИ-ассистент: вопросы по своим показателям и подготовка к визиту к врачу

### Чего сервис НЕ делает

- ❌ Не ставит диагнозы
- ❌ Не назначает и не отменяет препараты, не подбирает дозировки
- ❌ Не анализирует снимки (ЭКГ, УЗИ, КТ, МРТ, рентген)
- ❌ Не даёт рекомендаций по лечению
- ❌ Не добавляет в сводку значения без подтверждения пользователем
- ❌ Не является медицинским заключением

---

## ✅ Что уже сделано

Реализованы все ключевые требования задания (FR1–FR7), а также ИИ-обзор сводки
и ИИ-ассистент (FR8/FR10/FR15): бэкенд на Go (REST API) и переработанный
фронтенд на Next.js.

| Требование | Статус | Где |
| --- | --- | --- |
| **FR1** — регистрация, вход, выход, согласие на обработку данных | ✅ | `backend/internal/app/auth.go`, `consent.go`; `frontend/app/{login,register,consent}` |
| **FR2** — анкета здоровья (пол, дата рождения, анамнез, факторы риска) | ✅ | `backend/internal/app/profile.go`; `frontend/app/profile` |
| **FR3** — диагнозы и лекарства (CRUD) | ✅ | `backend/internal/app/`; вкладки в `frontend/app/profile` |
| **FR4** — домашние измерения: давление, пульс, вес | ✅ | `backend/internal/app/measurements.go`; `frontend/app/profile` |
| **FR5** — загрузка бланков (PDF, JPG, JPEG, PNG, DOCX) | ✅ | `backend/internal/{docparse,storage}`, `app/documents.go`; `frontend/app/analysis` |
| **FR6** — распознавание показателей через ИИ-vision | ✅ | `backend/internal/ai` (по умолчанию **mock**; OpenRouter/Gemini/Claude — опционально); `guardrails.go` |
| **FR7** — ручная проверка, правка, удаление и подтверждение показателей | ✅ | `backend/internal/app/indicators.go`; таблица проверки в `frontend/app/analysis` |
| **FR8** — отклонения подтверждённых показателей от референса | ✅ | `backend/internal/app/summary.go` (`classifyDeviation`); референс берётся из бланка, при отсутствии — типовой (`ai.DefaultReference`); `frontend/app/summary` |
| **FR10** — личная сводка (свод данных) | ✅ | `GET /api/summary`; `frontend/app/summary` |
| **FR15** — ИИ-обзор и чат-ассистент по своим данным | ✅ | `POST /api/summary/overview`, `POST /api/chat`; `frontend/components/custom/{AiOverviewCard,AiChat}.tsx` |

Дополнительно:

- **Безопасность**: cookie-сессии (`gorilla/sessions`) + `bcrypt`; все записи
  привязаны к `user_id`; в сводку попадают только значения со статусом
  `confirmed`/`manual`; guardrails (prompt + серверный пост-фильтр) не допускают
  диагнозов, назначений и трактовки снимков.
- **ИИ-обзор и чат**: отклонения от референса считаются детерминированно на
  сервере (FR8), ИИ добавляет связный текст; ответы проходят серверный
  пост-фильтр guardrails, в контекст попадают только подтверждённые значения.
- **Публичный лендинг** AI MED с видео-фоном, анимациями и адаптивом
  (`frontend/components/custom/Landing.tsx`, `landing.css`).
- **Переключатель светлой/тёмной темы** во всём приложении
  (`frontend/components/custom/ThemeToggle.tsx`, токены `.dark` в `globals.css`).
- **Единый дизайн** (clinical Minimalism): токены, статус-чипы, переработанные
  экраны входа, дашборда, анализов и профиля.
- **Код и комментарии** на русском; `gofmt`/`go vet`/`tsc`/`next build` проходят
  без ошибок.

**ИИ-режим:** по умолчанию `AI_PROVIDER=mock` — все ИИ-функции (распознавание
FR6/FR7, обзор и чат) работают на встроенных демо-ответах без сетевых вызовов и
ключей, что удобно для показа MVP. Для реального распознавания задайте ключ и
`AI_PROVIDER=openrouter` (или `gemini`/`anthropic`) — код менять не нужно.

---

## 🛠 Технологии

### Frontend

- **React 19.1.0**, **Next.js 15.4.5** (App Router), **TypeScript 5**
- **Tailwind CSS 4**, **Lucide React**, **Recharts**

### Backend

- **Go 1.26** — REST API (`chi/v5`)
- **PostgreSQL** через `pgx/v5` (`pgxpool`)
- **Аутентификация**: cookie-сессии (`gorilla/sessions`) + `bcrypt`
- **ИИ-провайдер**: по умолчанию **mock** (демо без ключей); опционально любой
  **OpenAI-совместимый** API (**OpenRouter** и др.), **Gemini** и
  **Anthropic Claude** — как резервные; вызовы через штатный `net/http` без SDK.
  PDF растеризуется в PNG (`pdftoppm`, poppler) до отправки в vision-модель.

---

## 🚀 Быстрый старт

### Требования

- Node.js 18+
- Go 1.26+
- PostgreSQL 16+
- `poppler-utils` (`pdftoppm`) — растеризация PDF для OCR

### Frontend

```bash
cd frontend
npm install
npm run dev
# http://localhost:3000
```

Запросы к `/api/*` проксируются на backend `http://localhost:5001`
(см. `frontend/next.config.ts`).

### Backend

```bash
cd backend
cp .env.example .env      # заполните DATABASE_URL и ключи ИИ
go run ./cmd/server
```

Сервер при старте применяет SQL-миграции (встроены через `embed`) и слушает
порт `5001`.

#### Переменные окружения (основные)

| Переменная | Назначение |
| --- | --- |
| `DATABASE_URL` | DSN PostgreSQL (или `DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD`) |
| `SECRET_KEY` | Ключ подписи cookie-сессий |
| `UPLOAD_DIR` | Каталог для загруженных оригиналов (по умолчанию `uploads`) |
| `CORS_ORIGINS` | Разрешённые origin'ы фронтенда |
| `PORT` | Порт API (по умолчанию `5001`) |
| `AI_PROVIDER` | `mock` (по умолчанию, демо без ключей), `openrouter`, `gemini` или `anthropic` |
| `OPENAI_API_KEY` | Ключ OpenAI-совместимого API (алиас `OPENROUTER_API_KEY`) |
| `OPENAI_BASE_URL` | Базовый URL (по умолчанию `https://openrouter.ai/api/v1`) |
| `OPENAI_MODEL` | Модель(и) через запятую — перебираются по порядку при лимитах |
| `GEMINI_API_KEY` / `GEMINI_MODEL` | Доступ к Gemini (модель по умолчанию `gemini-3.8-flash`) |
| `ANTHROPIC_API_KEY` / `ANTHROPIC_MODEL` | Доступ к Claude |

Провайдер выбирается по `AI_PROVIDER`: `mock` (по умолчанию) отвечает из
встроенных демо-данных без сетевых вызовов; для `openrouter`/`gemini`/`anthropic`
ключи подключаются по наличию, а остальные становятся резервом в порядке
`openrouter → gemini → anthropic`. `OPENAI_MODEL` можно задать списком через
запятую (например, несколько бесплатных моделей OpenRouter) — они перебираются по
порядку при лимитах.

Без ключей ИИ сервис продолжает работать: загрузка сохраняется, но распознавание
показателей завершится статусом «Ошибка».

---

## 🔌 API (основное)

| Метод | Путь | Назначение |
| --- | --- | --- |
| `POST` | `/api/auth/register` | Регистрация |
| `POST` | `/api/auth/login` | Вход |
| `POST` | `/api/auth/logout` | Выход |
| `GET` | `/api/auth/me` | Текущий пользователь + наличие согласия |
| `GET`/`POST` | `/api/consent` | Согласие на обработку данных |
| `GET`/`PUT` | `/api/profile/questionnaire` | Анкета здоровья (FR2) |
| `GET`/`POST` | `/api/diagnoses`, `PUT`/`DELETE` `/api/diagnoses/{id}` | Диагнозы (FR3) |
| `GET`/`POST` | `/api/medications`, `PUT`/`DELETE` `/api/medications/{id}` | Лекарства (FR3) |
| `GET`/`POST` | `/api/measurements`, `PUT`/`DELETE` `/api/measurements/{id}` | Измерения (FR4) |
| `GET`/`POST` | `/api/documents`, `GET`/`DELETE` `/api/documents/{id}` | Документы (FR5) |
| `GET` | `/api/documents/{id}/file`, `/api/documents/{id}/status` | Оригинал и статус |
| `GET` | `/api/documents/{id}/indicators` | Распознанные показатели (FR6) |
| `PUT`/`DELETE` | `/api/indicators/{id}` | Правка/удаление показателя (FR7) |
| `POST` | `/api/indicators/{id}/confirm`, `/needs-review` | Статус проверки (FR7) |
| `POST` | `/api/documents/{id}/confirm-all` | Подтвердить все показатели |
| `GET` | `/api/summary` | Личная сводка: показатели, отклонения, измерения (FR10) |
| `POST` | `/api/summary/overview` | ИИ-обзор сводки и подготовка к визиту (FR8/FR15) |
| `POST` | `/api/chat` | ИИ-чат по своим подтверждённым данным (FR15) |

Все данные привязаны к пользователю: чужие записи недоступны.

---

## 🏗 Архитектура

```
ai-med/
├── frontend/                       # Next.js приложение
│   ├── app/
│   │   ├── page.tsx                # Обзор (дашборд) / лендинг для гостя
│   │   ├── summary/                # Личная сводка + ИИ-обзор и чат (FR8/FR10/FR15)
│   │   ├── login/ register/        # Вход и регистрация
│   │   ├── consent/                # Согласие на обработку данных
│   │   ├── analysis/               # Загрузка документов и проверка (FR5–FR7)
│   │   └── profile/                # Анкета, диагнозы, лекарства, измерения
│   ├── components/custom/          # AuthGate, AppShell, sidebar, AiOverviewCard, AiChat, ...
│   └── lib/api.ts                  # HTTP-хелпер
└── backend/                        # Go API
    ├── cmd/server/main.go          # точка входа
    └── internal/
        ├── app/                    # HTTP-хендлеры (auth, profile, documents, summary)
        ├── ai/                     # провайдеры mock/OpenRouter/Gemini/Anthropic + guardrails
        ├── db/                     # пул pgx + SQL-миграции
        ├── docparse/               # форматы и разбор DOCX
        ├── storage/                # локальное хранилище оригиналов
        ├── httpx/                  # JSON-хелперы
        └── config/                 # конфигурация из окружения
```


## 🔒 Безопасность и приватность

- Пароли хранятся только в виде `bcrypt`-хеша.
- Доступ к данным — по подписанной cookie-сессии; все записи привязаны к `user_id`.
- Оригиналы документов хранятся локально в каталоге пользователя.
- Распознанные значения получают статус «требует проверки» и попадают в сводку
  только после подтверждения.
- Сервис не является медицинским изделием и не заменяет консультацию врача.

---

## 📄 Лицензия

MIT — см. файл [LICENSE](LICENSE).
