CREATE TABLE misiones (
    id SERIAL PRIMARY KEY,
    titulo TEXT NOT NULL,
    categoria TEXT NOT NULL,
    dificultad TEXT NOT NULL,
    estado TEXT NOT NULL,
    created_at TIMESTAMP
);