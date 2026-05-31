-- Down migration
DROP TABLE IF EXISTS ad_events;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS ads;
DROP TABLE IF EXISTS ad_cpm_rates;
DROP TABLE IF EXISTS ad_types;
DROP TABLE IF EXISTS ad_categories;

ALTER TABLE urls DROP COLUMN IF EXISTS allowed_categories;
ALTER TABLE urls DROP COLUMN IF EXISTS is_monetized;
ALTER TABLE users DROP COLUMN IF EXISTS role;

DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS transaction_status;
