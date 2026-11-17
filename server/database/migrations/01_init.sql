CREATE TABLE IF NOT EXISTS tags (
    guild_id   BIGINT NOT NULL,
    owner_id   BIGINT NOT NULL,
    name       TEXT NOT NULL,
    content    TEXT NOT NULL,
    uses       INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (guild_id, name)
);

CREATE TABLE IF NOT EXISTS contributors (
    username      TEXT PRIMARY KEY,
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type    TEXT NOT NULL,
    scopes        JSONB NOT NULL,
    expiration    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS aliases (
    alias  TEXT PRIMARY KEY,
    module TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS github_releases (
    name          TEXT PRIMARY KEY,
    webhook_id    BIGINT NOT NULL,
    ping_role     BIGINT NOT NULL,
    webhook_token TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS contributor_repos (
    name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS mod_mail_threads (
    channel_id BIGINT PRIMARY KEY,
    thread_id  BIGINT NOT NULL UNIQUE
);
