CREATE TABLE IF NOT EXISTS consent_log (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id    UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    consent_type   TEXT        NOT NULL,
    ip_address     INET,
    user_agent     TEXT,
    terms_version  TEXT,
    consented_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_consent_log_customer ON consent_log(customer_id, consented_at DESC);
