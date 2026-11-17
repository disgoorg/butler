-- name: GetTag :one
SELECT guild_id, owner_id, name, content, uses, created_at, updated_at
FROM tags
WHERE guild_id = $1 AND name = $2;

-- name: GetAndIncrementTag :one
UPDATE tags
SET uses = uses + 1
WHERE guild_id = $1 AND name = $2
RETURNING content;

-- name: ListTags :many
SELECT guild_id, owner_id, name, content, uses, created_at, updated_at
FROM tags
WHERE guild_id = $1
ORDER BY name;

-- name: ListTagsByOwner :many
SELECT guild_id, owner_id, name, content, uses, created_at, updated_at
FROM tags
WHERE guild_id = $1 AND owner_id = $2
ORDER BY name;

-- name: CreateTag :exec
INSERT INTO tags (guild_id, owner_id, name, content)
VALUES ($1, $2, $3, $4);

-- name: UpdateTag :exec
UPDATE tags
SET content = $3, updated_at = now()
WHERE guild_id = $1 AND name = $2;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE guild_id = $1 AND name = $2;
