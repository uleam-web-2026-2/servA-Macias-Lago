-- name: BuscarMisionSqlc :one
SELECT id, titulo, categoria, dificultad, estado, usuario_id, created_at
FROM misions
WHERE id = $1 AND deleted_at IS NULL;

-- name: GuardarMisionSqlc :execrows
UPDATE misions
SET titulo = $2, categoria = $3, dificultad = $4, estado = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;