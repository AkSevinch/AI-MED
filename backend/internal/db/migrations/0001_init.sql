-- AI MED core schema (FR1-FR7)

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consents (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version     VARCHAR(50) NOT NULL,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_consents_user ON consents (user_id);

CREATE TABLE IF NOT EXISTS health_profiles (
    user_id                 INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    birth_date              DATE,
    sex                     VARCHAR(20),
    known_diagnoses         TEXT,
    current_medications     TEXT,
    cardiovascular_history  VARCHAR(20),
    weight_kg               NUMERIC,
    risk_factors            TEXT,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS diagnoses (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    comment    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_diagnoses_user ON diagnoses (user_id);

CREATE TABLE IF NOT EXISTS medications (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    drug_class VARCHAR(255),
    frequency  VARCHAR(255),
    status     VARCHAR(20) NOT NULL DEFAULT 'unknown',
    comment    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_medications_user ON medications (user_id);

CREATE TABLE IF NOT EXISTS measurements (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        VARCHAR(20) NOT NULL,
    measured_at DATE NOT NULL,
    systolic    INTEGER,
    diastolic   INTEGER,
    pulse       INTEGER,
    weight      NUMERIC,
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_measurements_user ON measurements (user_id);

CREATE TABLE IF NOT EXISTS lab_documents (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename    TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    mime        VARCHAR(100),
    status      VARCHAR(20) NOT NULL DEFAULT 'processing',
    lab_name    VARCHAR(255),
    test_date   DATE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_lab_documents_user ON lab_documents (user_id);

CREATE TABLE IF NOT EXISTS lab_indicators (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_id         INTEGER REFERENCES lab_documents(id) ON DELETE SET NULL,
    code                VARCHAR(50),
    name                VARCHAR(255) NOT NULL,
    value               VARCHAR(100),
    unit                VARCHAR(50),
    reference_range     VARCHAR(255),
    test_date           DATE,
    source              VARCHAR(10) NOT NULL DEFAULT 'ocr',
    verification_status VARCHAR(20) NOT NULL DEFAULT 'needs_review',
    confidence          NUMERIC,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_lab_indicators_user ON lab_indicators (user_id);
CREATE INDEX IF NOT EXISTS idx_lab_indicators_document ON lab_indicators (document_id);
