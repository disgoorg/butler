-- name: ListModMailThreads :many
SELECT channel_id, thread_id
FROM mod_mail_threads
ORDER BY channel_id;

-- name: CreateModMailThread :exec
INSERT INTO mod_mail_threads (channel_id, thread_id)
VALUES ($1, $2);

-- name: DeleteModMailThread :exec
DELETE FROM mod_mail_threads
WHERE channel_id = $1 OR thread_id = $1;
