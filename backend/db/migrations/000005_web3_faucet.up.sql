-- Up migration
-- Track faucet claims per user (one claim per 24h per account)
CREATE TABLE IF NOT EXISTS faucet_claims (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id),
    amount      DECIMAL(42,0) NOT NULL,
    tx_hash     VARCHAR(66) UNIQUE,
    claimed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_faucet_claims_user ON faucet_claims(user_id);
CREATE INDEX idx_faucet_claims_claimed_at ON faucet_claims(claimed_at);
