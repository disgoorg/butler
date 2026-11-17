-- name: ListContributors :many
SELECT username, access_token, refresh_token, token_type, scopes, expiration
FROM contributors
ORDER BY username;

-- name: CreateContributor :exec
INSERT INTO contributors (username, access_token, refresh_token, token_type, scopes, expiration)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (username) DO UPDATE SET
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    token_type = EXCLUDED.token_type,
    scopes = EXCLUDED.scopes,
    expiration = EXCLUDED.expiration;

-- name: DeleteContributor :exec
DELETE FROM contributors
WHERE username = $1;
