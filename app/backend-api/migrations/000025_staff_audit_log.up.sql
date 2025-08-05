-- ==================== Staff Audit Log ====================

CREATE TABLE IF NOT EXISTS staff_audit_log (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id       UUID        NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    staff_id      UUID        REFERENCES shop_staff(id) ON DELETE SET NULL,
    action        TEXT        NOT NULL,
    resource_type TEXT        NOT NULL DEFAULT '',
    resource_id   UUID,
    metadata      JSONB       NOT NULL DEFAULT '{}',
    ip_address    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staff_audit_log_shop   ON staff_audit_log(shop_id, created_at DESC);
CREATE INDEX idx_staff_audit_log_staff  ON staff_audit_log(staff_id, created_at DESC);
