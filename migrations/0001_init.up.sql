CREATE TABLE IF NOT EXISTS teams (
    team_name      TEXT PRIMARY KEY,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    user_id    TEXT PRIMARY KEY,
    username   TEXT        NOT NULL,
    team_name  TEXT        NOT NULL REFERENCES teams(team_name) ON UPDATE CASCADE,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(team_name, is_active);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT        NOT NULL,
    author_id         TEXT        NOT NULL REFERENCES users(user_id) ON UPDATE CASCADE,
    status            TEXT        NOT NULL CHECK (status IN ('OPEN','MERGED')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author ON pull_requests(author_id);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id TEXT    NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT    NOT NULL REFERENCES users(user_id) ON UPDATE CASCADE,
    slot            SMALLINT NOT NULL CHECK (slot BETWEEN 1 AND 2),
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, slot),
    UNIQUE (pull_request_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_reviewers_reviewer ON pull_request_reviewers(reviewer_id, pull_request_id);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key_hash      BYTEA PRIMARY KEY,
    method        TEXT        NOT NULL,
    path          TEXT        NOT NULL,
    request_body  BYTEA       NOT NULL,
    response_body BYTEA       NOT NULL,
    response_headers JSONB    NOT NULL DEFAULT '{}'::jsonb,
    status_code   INT         NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_idempotency_expiry ON idempotency_keys(expires_at);
