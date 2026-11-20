# Semana 12 · auditoría de inyección

Búsqueda: `git grep -n -E -f docs/taller12/patrones.txt -- "*.go" ":(exclude)*_test.go"`. Salida completa en `auditoria.txt` (21 líneas).

| Línea | ¿Arma SQL? | ¿Llega texto del cliente? | ¿Va aparte? | Qué hicimos |
| --- | --- | --- | --- | --- |
| `internal/misiones/manejadores.go:100` y `203` (`estado := r.URL.Query().Get("estado")`) | No (lee query param) | Sí, `?estado=` | Pasa como argumento a repositorio | Nada: solo lee el parámetro de consulta. |
| `internal/misiones/repositorio.go:43` (`a.DB.Find(&paquetes).Error`) | Sí | No | — | Nada: consulta fija sin entrada del cliente. |
| `internal/misiones/repositorio.go:51` y `98` (`q = q.Where("estado = ?", estado)`) | Sí | Sí, `?estado=` | Sí, con `?` en GORM | Nada: es segura, el valor viaja parametrizado. |
| `internal/misiones/repositorio.go:53` y `100` (`q.Order("id asc").Find(&pagos)` / `Find(&lista)`) | Sí | No (`"id asc"` es fijo) | — | Nada: orden fijo en código, sin columna enviada por el cliente. |
| `internal/misiones/repositorio.go:63` y `106` (`a.DB.First(&p, id).Error` / `First(&m, id).Error`) | Sí | Sí, `{id}` de ruta | Sí, `id` ya convertido en transporte | Nada: es segura. |
| `internal/misiones/sqlc/db.go:13, 15, 16` (`ExecContext`, `QueryContext`, `QueryRowContext`) | Interfaz de `sqlc` | Según llamada | Sí, argumentos parametrizados | Nada: interfaz generada por `sqlc`. |
| `internal/misiones/sqlc/query.sql.go:20, 50, 76, 97, 126` (`buscarPagoSqlc`, `guardarPagoSqlc`, `crearFacturaSqlc`, `buscarMisionSqlc`, `guardarMisionSqlc`) | Sí | Sí (`id`, `arg.*`) | Sí, con `$1`, `$2`... en `sqlc` | Nada: código generado por `sqlc` con consultas parametrizadas. |
| `internal/sesion/guardia.go:60` (`"solo el rol "+rol+" puede hacer esto"`) y `main.go:64` (`Addr: ":" + cfg.Puerto`) | No (mensaje de error y dirección del servidor) | No | — | Nada: no arman SQL. |
| `internal/sesion/usuarios.go:41` y `67` (`Where("nombre = ?", ...).Limit(1).Find(&u)`) | Sí | Sí (`e.nombre` / `pedido.Usuario`) | Sí, con `?` en GORM | Nada: es segura, el valor viaja como parámetro. |

**Resultado:** Las 21 líneas revisadas no concatenan entrada del cliente dentro de sentencias SQL con `+` ni `Sprintf`; las consultas usan parámetros (`?` en GORM o argumentos en `sqlc`), y en `inyeccion.txt` el filtro normal devolvió `200 OK` con el registro `"pendiente"` mientras `' OR 1=1 --` devolvió `200 OK` con `[]` sin traer registros adicionales.