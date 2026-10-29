# Hoja del Hito 2 - Double Level

## 1. Usuarios de ejemplo
| Usuario | Rol | Contraseña de ejemplo |
| :--- | :--- | :--- |
| `ana` | jugador | `ana-2026` |
| `luis` | admin | `luis-2026` |

## 2. Las seis peticiones
| # | Petición | Con qué credencial | Debe responder |
| :-: | :--- | :--- | :---: |
| 1 | `GET /pagos` | Ninguna | `401` |
| 2 | `POST /login` (con ana) | Ninguna | `200` y el token |
| 3 | `GET /pagos` | Token de Ana | `200` |
| 4 | `PATCH /pagos/{id}` → `pago_por_verificar` | Token de Ana (jugador) | `403` |
| 5 | `PATCH /pagos/{id}` → `pago_por_verificar` | Token de Luis (admin) | `200` |
| 6 | `PATCH /pagos/1` → `completado`, estando `pendiente` | Token de Luis (admin) | `422` |

## 3. Matriz ruta x rol
| Ruta | Sin sesión | Jugador (Ana) | Admin (Luis) |
| :--- | :---: | :---: | :---: |
| `POST /login` | Sí | Sí | Sí |
| `GET /paquetes` | Sí | Sí | Sí |
| `GET /pagos` | No (401) | Sí | Sí |
| `POST /pagos` | No (401) | Sí | Sí |
| `GET /pagos/{id}` | No (401) | Sí | Sí |
| `PUT /pagos/{id}` | No (401) | No (403) | Sí |
| `PATCH /pagos/{id}` | No (401) | No (403) | Sí |
| `DELETE /pagos/{id}` | No (401) | No (403) | Sí |

## 4. Recuadro de cumplimiento
- [x] El validador termina sin errores ni avisos (`npx @redocly/cli lint`).
- [x] Los structs de todas las entidades están en el código y sus tablas existen en la base de datos.
- [x] Las seis peticiones están en la colección de Bruno, en orden.
- [x] Cada una responde lo que dice esta hoja.
- [x] `go test ./...` en verde, con una prueba de integración.
- [x] El código de la entidad principal está en capas y la estructura está explicada en `docs/decisiones.md`.
- [x] Hay una captura de respaldo de cada paso en `docs/hito2_capturas/`.