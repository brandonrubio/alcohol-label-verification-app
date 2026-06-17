CREATE TABLE IF NOT EXISTS verification_batches (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'processing',
    total_count INTEGER NOT NULL DEFAULT 0,
    completed_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS verifications (
    id UUID PRIMARY KEY,
    batch_id UUID REFERENCES verification_batches(id) ON DELETE SET NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    image_name TEXT NOT NULL,
    application_json JSONB NOT NULL,
    extracted_json JSONB,
    field_results_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    processing_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verifications_user_id ON verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_verifications_batch_id ON verifications(batch_id);
CREATE INDEX IF NOT EXISTS idx_verification_batches_user_id ON verification_batches(user_id);
