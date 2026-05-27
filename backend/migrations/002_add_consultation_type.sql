-- Add consultation_type to distinguish new designs vs makeup/enhance/cover-up
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
END $$;

-- Add forensic analysis columns for existing tattoo (makeup mode)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'consultations' AND column_name = 'tattoo_analysis'
    ) THEN
        ALTER TABLE consultations ADD COLUMN tattoo_analysis TEXT;
    END IF;
END $$;
