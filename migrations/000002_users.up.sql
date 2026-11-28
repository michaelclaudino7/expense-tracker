CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Every category now belongs to a user.
ALTER TABLE categories
    ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

-- Names are unique per user, not globally.
ALTER TABLE categories DROP CONSTRAINT categories_name_type_key;
ALTER TABLE categories
    ADD CONSTRAINT categories_user_name_type_key UNIQUE (user_id, name, type);

-- Every transaction now belongs to a user.
ALTER TABLE transactions
    ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_transactions_user_id ON transactions (user_id);