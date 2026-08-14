CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    session_id VARCHAR(255) NOT NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    url_path VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analytics_events_shop_created ON analytics_events(shop_id, created_at);
CREATE INDEX idx_analytics_events_session_created ON analytics_events(session_id, created_at DESC);
