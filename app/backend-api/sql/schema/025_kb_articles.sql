CREATE TABLE IF NOT EXISTS kb_articles (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id       UUID        NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    category      TEXT        NOT NULL DEFAULT 'general',
    title         TEXT        NOT NULL,
    slug          TEXT        NOT NULL,
    body          TEXT        NOT NULL,
    is_published  BOOLEAN     NOT NULL DEFAULT false,
    position      INT         NOT NULL DEFAULT 0,
    view_count    BIGINT      NOT NULL DEFAULT 0,
    helpful_count BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(shop_id, slug)
);
CREATE INDEX idx_kb_articles_shop_cat ON kb_articles(shop_id, category, is_published);
