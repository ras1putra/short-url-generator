CREATE TABLE clicks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url_id      UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    ip_hash     VARCHAR(64),
    country     VARCHAR(2),
    city        VARCHAR(100),
    device      VARCHAR(10),
    browser     VARCHAR(20),
    referrer    TEXT,
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clicks_url_id ON clicks(url_id);
CREATE INDEX idx_clicks_clicked_at ON clicks(clicked_at);
CREATE INDEX idx_clicks_url_date ON clicks(url_id, clicked_at);
