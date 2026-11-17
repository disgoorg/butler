-- name: ListContributorRepos :many
SELECT name
FROM contributor_repos
ORDER BY name;

-- name: CreateContributorRepo :exec
INSERT INTO contributor_repos (name)
VALUES ($1);

-- name: DeleteContributorRepo :exec
DELETE FROM contributor_repos
WHERE name = $1;
