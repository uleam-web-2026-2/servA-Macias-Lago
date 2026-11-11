-- name: BuscarPagoSqlc :one
SELECT id, usuario_id, paquete_gemas_id, monto, estado, comprobante_url, fecha_creacion
FROM pagos
WHERE id = $1;

-- name: GuardarPagoSqlc :execrows
UPDATE pagos
SET usuario_id = $2, paquete_gemas_id = $3, monto = $4, estado = $5, comprobante_url = $6
WHERE id = $1;

-- name: CrearFacturaSqlc :exec
INSERT INTO facturas (pago_id, identificacion, razon_social, fecha_emision)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP);

-- name: BuscarMisionSqlc :one
SELECT id, titulo, categoria, dificultad, estado, usuario_id, created_at
FROM misions
WHERE id = $1 AND deleted_at IS NULL;

-- name: GuardarMisionSqlc :execrows
UPDATE misions
SET titulo = $2, categoria = $3, dificultad = $4, estado = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;