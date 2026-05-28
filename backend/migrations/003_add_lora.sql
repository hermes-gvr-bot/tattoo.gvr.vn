-- Add LoRA training support + multi-photo upload for body preview
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'loras') THEN
        CREATE TABLE loras (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
            status VARCHAR(20) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','uploading','training','ready','failed')),
            replicate_training_id VARCHAR(255),
            replicate_version_id VARCHAR(255),
            lora_weights_url TEXT,
            trigger_word VARCHAR(100),
            error_message TEXT,
            total_photos INT DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        );
        CREATE UNIQUE INDEX idx_loras_consultation ON loras(consultation_id);
        CREATE INDEX idx_loras_status ON loras(status);
    END IF;
END $$;

-- Store additional reference photos for LoRA training
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'consultation_photos') THEN
        CREATE TABLE consultation_photos (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
            photo_url TEXT NOT NULL,
            photo_order INT NOT NULL DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now()
        );
        CREATE INDEX idx_cons_photos_consultation ON consultation_photos(consultation_id);
    END IF;
END $$;

-- Add body_preview_path to variants for body-superimposed previews
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'variants' AND column_name = 'body_preview_path'
    ) THEN
        ALTER TABLE variants ADD COLUMN body_preview_path TEXT;
        ALTER TABLE variants ADD COLUMN body_preview_status VARCHAR(20) DEFAULT 'pending'
            CHECK (body_preview_status IN ('pending','generating','completed','failed'));
    END IF;
END $$;

-- Add consultation_type to status check (migration was already applied, this is idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'consultations' AND column_name = 'consultation_type'
    ) THEN
        ALTER TABLE consultations ADD COLUMN consultation_type VARCHAR(20) NOT NULL DEFAULT 'new_tattoo';
        ALTER TABLE consultations ADD CONSTRAINT chk_consultation_type
            CHECK (consultation_type IN ('new_tattoo', 'makeup_enhance'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'consultations' AND column_name = 'tattoo_analysis'
    ) THEN
        ALTER TABLE consultations ADD COLUMN tattoo_analysis TEXT;
    END IF;
END $$;
