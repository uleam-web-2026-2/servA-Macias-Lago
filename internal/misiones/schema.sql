CREATE TABLE misions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    titulo TEXT NOT NULL,
    categoria TEXT NOT NULL,
    dificultad TEXT NOT NULL,
    estado TEXT NOT NULL,
    usuario_id BIGINT NOT NULL DEFAULT 1
);