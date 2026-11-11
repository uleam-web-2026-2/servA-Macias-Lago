CREATE TABLE pagos (
    id BIGSERIAL PRIMARY KEY,
    usuario_id BIGINT NOT NULL,
    paquete_gemas_id BIGINT NOT NULL,
    monto NUMERIC(10,2) NOT NULL CHECK (monto > 0),
    estado VARCHAR(30) NOT NULL DEFAULT 'pendiente',
    comprobante_url TEXT,
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE facturas (
    id BIGSERIAL PRIMARY KEY,
    pago_id BIGINT NOT NULL UNIQUE,
    identificacion VARCHAR(20) NOT NULL CHECK (identificacion <> ''),
    razon_social VARCHAR(150) NOT NULL,
    fecha_emision TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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