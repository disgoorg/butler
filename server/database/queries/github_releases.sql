-- name: GetGithubRelease :one
SELECT name, webhook_id, ping_role, webhook_token
FROM github_releases
WHERE name = $1;

-- name: ListGithubReleases :many
SELECT name, webhook_id, ping_role, webhook_token
FROM github_releases
ORDER BY name;

-- name: CreateGithubRelease :exec
INSERT INTO github_releases (name, webhook_id, ping_role, webhook_token)
VALUES ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE SET
    webhook_id = EXCLUDED.webhook_id,
    ping_role = EXCLUDED.ping_role,
    webhook_token = EXCLUDED.webhook_token;

-- name: DeleteGithubRelease :exec
DELETE FROM github_releases
WHERE name = $1;
