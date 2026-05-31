ALTER TABLE users
  ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN email_verification_token VARCHAR(255),
  ADD COLUMN email_verification_sent_at TIMESTAMPTZ,
  ADD COLUMN password_reset_token VARCHAR(255),
  ADD COLUMN password_reset_sent_at TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_users_email_verification_token ON users(email_verification_token) WHERE email_verification_token IS NOT NULL;
CREATE UNIQUE INDEX idx_users_password_reset_token ON users(password_reset_token) WHERE password_reset_token IS NOT NULL;
