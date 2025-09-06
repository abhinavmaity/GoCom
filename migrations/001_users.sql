ALTER TABLE users ADD otp_secret VARCHAR(255);
ALTER TABLE users ADD otp_enabled BOOLEAN DEFAULT false;
ALTER TABLE users ADD otp_verified BOOLEAN DEFAULT false;

CREATE TABLE refresh_tokens (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
