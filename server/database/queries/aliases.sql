-- name: GetAlias :one
SELECT alias, module
FROM aliases
WHERE alias = $1;

-- name: ListAliases :many
SELECT alias, module
FROM aliases
ORDER BY alias;

-- name: CreateAlias :exec
INSERT INTO aliases (alias, module)
VALUES ($1, $2)
ON CONFLICT (alias) DO UPDATE SET module = EXCLUDED.module;

-- name: DeleteAlias :exec
DELETE FROM aliases
WHERE alias = $1;
