-- Tattoo Consultation Database Schema
-- Run: psql -h localhost -p 5437 -U tattoo -d tattoo_consultation -f migrations/001_init.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'client' CHECK (role IN ('client', 'admin')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- Consultations
-- ============================================================
CREATE TABLE IF NOT EXISTS consultations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    body_photo_path TEXT NOT NULL,
    idea_text TEXT NOT NULL,
    body_part VARCHAR(100),
    skin_tone VARCHAR(50),
    size_estimation VARCHAR(100),
    vision_analysis TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'pending_payment'
        CHECK (status IN ('pending_payment','payment_confirmed','generating','completed','delivered','rejected')),
    admin_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_consultations_user_id ON consultations(user_id);
CREATE INDEX IF NOT EXISTS idx_consultations_status ON consultations(status);

-- ============================================================
-- Design Variants (3 per consultation)
-- ============================================================
CREATE TABLE IF NOT EXISTS variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
    variant_number INT NOT NULL CHECK (variant_number BETWEEN 1 AND 3),
    prompt_used TEXT NOT NULL DEFAULT '',
    sketch_path TEXT,
    final_path TEXT,
    sketch_status VARCHAR(20) DEFAULT 'pending'
        CHECK (sketch_status IN ('pending','generating','completed','failed')),
    final_status VARCHAR(20) DEFAULT 'pending'
        CHECK (final_status IN ('pending','generating','completed','failed')),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(consultation_id, variant_number)
);

CREATE INDEX IF NOT EXISTS idx_variants_consultation ON variants(consultation_id);

-- ============================================================
-- Feedback & Iterations
-- ============================================================
CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
    selected_variant UUID REFERENCES variants(id),
    revision_notes TEXT NOT NULL,
    iteration_number INT NOT NULL DEFAULT 1,
    status VARCHAR(20) DEFAULT 'pending'
        CHECK (status IN ('pending','generating','completed')),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_consultation ON feedbacks(consultation_id);

-- ============================================================
-- Payments
-- ============================================================
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL REFERENCES consultations(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount_cents INT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USDT',
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('tatum','sepay')),
    txn_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','confirmed','failed','refunded')),
    created_at TIMESTAMPTZ DEFAULT now(),
    confirmed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payments_consultation ON payments(consultation_id);
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);

-- ============================================================
-- Seed Data
-- ============================================================
-- Admin account: x54@gvr.vn / Admin@2026!
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("Admin@2026!"), 12)
INSERT INTO users (email, password_hash, name, role)
VALUES ('x54@gvr.vn', '$2b$12$YrGaVTyPKDFygSL7VOgllev.JfvYXyUTwqOY5.U6itEeJOrefs7Di', 'Admin', 'admin')
ON CONFLICT (email) DO NOTHING;
