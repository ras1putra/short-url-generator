-- Up migration
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'advertiser', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'user';

-- Ad categories with display labels and pricing multipliers
CREATE TABLE ad_categories (
    category   VARCHAR(50) PRIMARY KEY,
    label      VARCHAR(100) NOT NULL,
    multiplier DECIMAL(5,2) NOT NULL DEFAULT 1.00
);

-- Ad types configuration
CREATE TABLE ad_types (
    ad_type VARCHAR(50) PRIMARY KEY,
    label VARCHAR(100) NOT NULL,
    aspect_ratio DECIMAL(5,2) NOT NULL,
    recommended_resolution VARCHAR(50) NOT NULL
);

-- CPM rates
CREATE TABLE ad_cpm_rates (
    ad_type VARCHAR(50) PRIMARY KEY REFERENCES ad_types(ad_type) ON DELETE CASCADE,
    cpm DECIMAL(20,8) NOT NULL
);

INSERT INTO ad_categories (category, label, multiplier) VALUES
    ('regular',  'Regular',  1.00),
    ('crypto',   'Crypto',   1.00),
    ('gambling', 'Gambling', 2.00),
    ('adult',    'Adult',    3.00);

INSERT INTO ad_types (ad_type, label, aspect_ratio, recommended_resolution) VALUES
    ('BANNER', 'Banner / Display', 5.00, '1200×240'),
    ('NATIVE', 'Native Block', 1.20, '600×500'),
    ('POPUP', 'Pop-Up Ad', 1.33, '800×600'),
    ('VIDEO', 'Featured Slot', 1.78, '1920×1080'),
    ('INTERSTITIAL', 'Interstitial Overlay', 0.56, '1080×1920');

INSERT INTO ad_cpm_rates (ad_type, cpm) VALUES
    ('BANNER', 1.00),
    ('NATIVE', 1.50),
    ('POPUP', 3.00),
    ('VIDEO', 4.00),
    ('INTERSTITIAL', 5.00);

-- Ads table
CREATE TABLE ads (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advertiser_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    image_url         TEXT NOT NULL,
    target_url        TEXT NOT NULL,
    category          VARCHAR(50) NOT NULL REFERENCES ad_categories(category),
    ad_type           VARCHAR(50) NOT NULL DEFAULT 'BANNER' REFERENCES ad_types(ad_type),
    total_budget      DECIMAL(20,8) NOT NULL DEFAULT 0.00,
    remaining_budget  DECIMAL(20,8) NOT NULL DEFAULT 0.00,
    cpm               DECIMAL(20,8) NOT NULL DEFAULT 1.00,
    status            VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wallets table
CREATE TABLE wallets (
    user_id     UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance     DECIMAL(20,8) NOT NULL DEFAULT 0.00,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Transactions table
CREATE TABLE transactions (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      DECIMAL(20,8) NOT NULL,
    type        VARCHAR(20) NOT NULL, -- 'EARNING', 'AD_SPEND', 'DEPOSIT', 'WITHDRAWAL'
    tx_hash     VARCHAR(66) UNIQUE,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ad Events table
CREATE TABLE ad_events (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ad_id             UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
    link_id           UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    event_type        VARCHAR(20) NOT NULL, -- 'IMPRESSION', 'CLICK', 'COMPLETION'
    is_valid          BOOLEAN NOT NULL DEFAULT TRUE,
    quality_score     DECIMAL(3,2) NOT NULL DEFAULT 1.00,
    rejection_reason  VARCHAR(255),
    ip_address        INET,
    user_agent        TEXT,
    metadata          JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add monetization to urls
ALTER TABLE urls ADD COLUMN is_monetized BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE urls ADD COLUMN allowed_categories TEXT[];

-- Indexes
CREATE INDEX idx_ads_advertiser_id ON ads(advertiser_id);
CREATE INDEX idx_ads_status ON ads(status);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_ad_events_ad_id ON ad_events(ad_id);
CREATE INDEX idx_ad_events_link_id ON ad_events(link_id);
CREATE INDEX idx_urls_is_monetized ON urls(is_monetized);

-- Create wallets for existing users
INSERT INTO wallets (user_id, balance) SELECT id, 0.00 FROM users ON CONFLICT DO NOTHING;
