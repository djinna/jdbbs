-- name: GetProject :one
SELECT * FROM projects WHERE id = ?;

-- name: GetProjectByPath :one
SELECT * FROM projects WHERE client_slug = ? AND project_slug = ? AND archived_at IS NULL;

-- name: ListProjects :many
SELECT * FROM projects WHERE archived_at IS NULL ORDER BY updated_at DESC;

-- name: ListArchivedProjects :many
SELECT * FROM projects WHERE archived_at IS NOT NULL ORDER BY archived_at DESC, updated_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, start_date, client_slug, project_slug, created_at, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: UpdateProject :exec
UPDATE projects SET name = ?, start_date = ?, client_slug = ?, project_slug = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: ListTasks :many
SELECT * FROM tasks WHERE project_id = ? ORDER BY sort_order;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: CreateTask :one
INSERT INTO tasks (
    project_id, sort_order, assignee, title, is_milestone,
    orig_weeks, curr_weeks, orig_due, curr_due, actual_done, status,
    words, words_per_hour, hours, rate, budget_notes,
    orig_budget, curr_budget, actual_budget,
    created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateTask :exec
UPDATE tasks SET
    assignee = ?, title = ?, is_milestone = ?,
    orig_weeks = ?, curr_weeks = ?,
    orig_due = ?, curr_due = ?, actual_done = ?, status = ?,
    words = ?, words_per_hour = ?, hours = ?, rate = ?,
    budget_notes = ?, orig_budget = ?, curr_budget = ?, actual_budget = ?,
    sort_order = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: CreateAuthToken :exec
INSERT INTO auth_tokens (project_id, token_hash, label)
VALUES (?, ?, ?);

-- name: GetAuthToken :one
SELECT * FROM auth_tokens WHERE project_id = ? AND token_hash = ?;

-- name: ListAuthTokens :many
SELECT * FROM auth_tokens WHERE project_id = ?;

-- name: DeleteAuthTokensByProject :exec
DELETE FROM auth_tokens WHERE project_id = ?;
