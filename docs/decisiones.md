# Decisiones

## D1 · ¿Quién pone el estado al crear?
**Opciones:** que lo mande quien crea, o que lo ponga el servidor.
**Qué elegimos:** Que el servidor fije siempre el estado inicial `"pendiente"` al crear un `Pago` (y obtenga el `UsuarioID` del token de sesión), ignorando cualquier estado o usuario enviado en el JSON.
**Por qué, en nuestro negocio:** Porque ningún jugador puede crear un pago directamente en `"completado"` ni registrar pagos a nombre de otro jugador.
**Qué pasaría con la otra opción:** Si el cliente pudiera enviar el estado al crear, podría saltarse la verificación administrativa del comprobante bancario.

# Registro de Decisiones de Arquitectura

## Semana 4

### D1 · Si falta una variable al arrancar
**Decisión:** `DATABASE_URL` y `CLAVE_FIRMA` son variables obligatorias sin valor por defecto que detienen la ejecución si no están presentes. `PUERTO` y `TIEMPO_ESPERA_SEGUNDOS` toman `8080` y `5` respectivamente.
**Por qué:** En *DoubleLevel*, la cadena de conexión y la clave de firma JWT contienen secretos que cambian según el entorno. El puerto y timeout no contienen secretos y facilitan el arranque local rápido.
**Qué descartamos:** Mantener credenciales en `main.go` o predefinir una contraseña por defecto en el código fuente.

### D2 · Qué hicimos con la contraseña que ya está en el historial
**Decisión:** Se ejecutó un cambio de contraseña en PostgreSQL (`ALTER USER`) y se aisló el valor sensible en el archivo `.env` mediante la variable `DATABASE_URL`.
**Por qué:** Cualquier credencial presente en el historial de Git está expuesta. Cambiar la clave activa invalida los valores presentes en los commits anteriores.
**Qué descartamos:** Reescribir el historial de Git sin modificar la clave real en la base de datos.

### D3 · La prueba que no escribimos
**Prueba que no escribimos:** Creación exitosa que responda HTTP 201 Created contra base de datos.
**Dónde empieza la base:** En `internal/misiones/repositorio.go`, cuando se invoca a GORM/sqlc.
**Qué haría falta para probarla:** Se requiere un repositorio falso que implemente la interfaz `Repositorio` o levantar una instancia de PostgreSQL de prueba.

## Semana 8 Estructura · Arquitectura en Capas

**Estructura elegida:** Separación por responsabilidades dentro del paquete de dominio `internal/misiones/` en tres capas: Transporte HTTP (`manejadores.go` y `dtos.go`), Reglas del Negocio (`servicio.go` y `estado.go`) y Repositorio de Datos (`repositorio.go` y `modelos.go`).

**Qué va en cada archivo y por qué:**
- **`manejadores.go` y `dtos.go` (Capa HTTP / Transporte):** Recibe la petición HTTP, decodifica los DTOs de entrada (`PagoPeticion`, `MisionPeticion`), extrae el `UsuarioID` del token JWT, invoca al `Servicio` y responde con DTOs de salida (`PagoRespuesta`, `MisionRespuesta`). **No consulta la base de datos directamente.**
- **`servicio.go` y `estado.go` (Capa de Negocio / Dominio):** Contiene la máquina de estados de `Pago` (`siguientes`), `puedePasar`, `ValidarCambioEstado` y los casos de uso. **No importa `net/http`, ni `gorm`, ni `database/sql`**.
- **`repositorio.go` y `modelos.go` (Capa de Persistencia / Datos):** Define la interfaz `Repositorio` e implementa los dos adaptadores intercambiables: `AlmacenGorm` y `PagosSqlc`.

## Semana 10 — transacciones

### D1 · El caso
Implementamos en el código el caso de uso **`ConfirmarPagoYEmitirFactura(id, identificacion, razonSocial)`** en `internal/misiones/servicio.go`. Cuando el administrador aprueba un pago que está en `"pago_por_verificar"`, el sistema realiza dos escrituras que deben ir juntas: actualiza el estado en la tabla `pagos` a `"completado"` y crea el comprobante fiscal en la tabla `facturas` (relación 1 a 1 de nuestra Ficha del Negocio). Si la segunda escritura falla (por ejemplo, cuando `identificacion` llega vacía y la base o el repositorio la rechaza por la restricción `CHECK (identificacion <> '')`), sin transacción el pago quedaría en estado `"completado"` con `0` facturas emitidas.

### D2 · Opciones que consideramos
1. **Opción A (Manejar `*gorm.DB` o `*sql.Tx` dentro de `Servicio`):** Hacer que `servicio.go` abra `Begin()`, ejecute los dos pasos y llame a `Commit()` o `Rollback()`.
   - *Qué cumple o rompe:* Deshace el fallo en PostgreSQL, pero **rompe la Restricción 1** (el servicio importaría `gorm` o `database/sql`) y **rompe la Restricción 2** (las pruebas unitarias con repositorio falso sin base dejarían de compilar).
2. **Opción B (Método atómico `GuardarPagoYFactura(p *Pago, f *Factura) error` en la interfaz `Repositorio`):** El servicio valida la transición de estado con `ValidarCambioEstado` y delega las dos escrituras juntas al método `GuardarPagoYFactura` de la interfaz `Repositorio`. Cada adaptador (`AlmacenGorm` y `PagosSqlc`) resuelve su propia transacción internamente.
   - *Qué cumple o rompe:* **Cumple las 5 restricciones**: `servicio.go` no importa `gorm` ni `database/sql`, las pruebas corren sin base usando `repoFalsoPago`, los manejadores no cambian, funciona en ambos adaptadores y deshace todo si falla el segundo paso.

### D3 · La que elegimos y por qué
Elegimos e implementamos la **Opción B (`GuardarPagoYFactura` en la interfaz `Repositorio`)**. En **DoubleLevel**, esto permite que `ConfirmarPagoYEmitirFactura` en `servicio.go` valide primero que el pago pueda pasar a `"completado"` y luego invoque una sola operación atómica del repositorio, manteniendo `servicio.go` libre de imports de base de datos y probado en `manejadores_test.go` (`TestTransaccionErrorForzadoDeshaceCambioDeEstado`).

### D4 · La que descartamos y por qué
Descartamos la **Opción A** porque acoplaría `servicio.go` a GORM o `database/sql` e impediría cambiar de adaptador en `main.go` sin tocar el servicio. También descartamos revertir el estado manualmente con un segundo `UPDATE` en Go porque si el servidor se cae entre el fallo de la factura y el `UPDATE` de reversión, el pago quedaría permanentemente en `"completado"` sin factura.

### D5 · Con sqlc
En `internal/misiones/repositorio.go`, dentro de `PagosSqlc.GuardarPagoYFactura`, abrimos la transacción sobre `*sql.DB` con `tx, err := a.Conexion.Begin()`, aseguramos `defer tx.Rollback()`, envolvemos las consultas generadas por `sqlc` con `qtx := a.q.WithTx(tx)` y ejecutamos `qtx.GuardarPagoSqlc` seguido de `qtx.CrearFacturaSqlc`. Si `CrearFacturaSqlc` devuelve error, la función retorna el error y `defer tx.Rollback()` deshace el cambio de estado del pago; solo si ambas consultas tienen éxito se ejecuta `tx.Commit()`. No hubo que cambiar nada en `servicio.go` ni en `manejadores.go`.

### D6 · Archivos
- **Archivos tocados:** `sqlc.yaml`, `internal/misiones/schema.sql`, `internal/misiones/query.sql`, `internal/misiones/sqlc/*`, `internal/misiones/modelos.go`, `internal/misiones/dtos.go`, `internal/misiones/repositorio.go`, `internal/misiones/servicio.go`, `internal/misiones/manejadores.go`, `internal/misiones/manejadores_test.go`, `main.go`, `docs/openapi.yaml`, `.env.example`, `docs/taller10/*` y `docs/decisiones.md`.
- **Archivos que quedaron igual:** `internal/config/config.go`, `internal/sesion/*`, `internal/respuesta/respuesta.go`, `internal/misiones/estado.go` y `internal/misiones/semilla.go`.


## Semana 12 · la seguridad de la API

### D1 · Lo que encontramos y cerramos
La búsqueda con `git grep` devolvió **21 líneas** en `docs/taller12/auditoria.txt`. De ellas, 11 arman consultas SQL donde llega entrada del cliente (los filtros `Where("estado = ?", estado)` en `repositorio.go`, `Where("nombre = ?", ...)` en `usuarios.go`, `First(&p, id)` con `id` ya convertido a `uint` mediante `strconv.Atoi`, y las consultas generadas por `sqlc` con `$1`, `$2`, etc.). No encontramos ninguna concatenación de entrada del cliente con `+` ni `Sprintf` dentro de SQL, por lo que no hubo que cerrar consultas vulnerables; todas envían el valor aparte como parámetro. En la prueba del paso 1.4 (`inyeccion.txt`), `GET /pagos?estado=pendiente` devolvió `200 OK` con el pago pendiente (`id: 1`), mientras que `GET /pagos?estado=' OR 1=1 --` devolvió `200 OK` con `[]` sin exponer registros adicionales.

### D2 · El origen que puede leer nuestras respuestas
Configuramos `AllowedOrigins: []string{"http://localhost:5173"}` (el origen de Vite en desarrollo) y no `"*"` por principio de mínimo privilegio, evitando que cualquier sitio externo pueda leer respuestas desde el navegador del usuario. Colocamos `cors.Handler` arriba de todo en `main.go` (antes de `middleware.Logger` y de cualquier ruta o guardia JWT) porque las peticiones previas `OPTIONS` del navegador no llevan token `Authorization` ni tienen manejador propio; si pasaran por el guardia responderían `401` o `405`. Al probar con `curl` un `PATCH /pagos/3` enviando `Origin: http://malo.example` y un token válido (`cors_curl_ejecuta.txt`), el servidor respondió `200 OK` y ejecutó el cambio, demostrando que CORS solo lo aplica el navegador y no sustituye la autenticación ni los roles. Cuando publiquemos el front, deberemos reemplazar o agregar el dominio real de producción en `AllowedOrigins` (idealmente desde una variable de entorno).

### D3 · El límite del login
Elegimos un tope de **5 intentos por minuto por IP** (`httprate.LimitBy(5, time.Minute, porIP)`), suficiente para que un usuario legítimo corrija un error de tipeo pero estricto frente a ataques de fuerza bruta sobre contraseñas. Se aplicó únicamente en `POST /login` con `r.With(...)` para no bloquear el uso normal de las demás rutas protegidas (como listar o actualizar pagos) cuando el cliente hace varias peticiones seguidas. En `rafaga.txt`, el primer `429 Too Many Requests` llegó en la **sexta petición** y trajo la cabecera `Retry-After: 60`.

### D4 · Detrás de un proxy
`middleware.ClientIPFromRemoteAddr` obtiene la IP directamente de `r.RemoteAddr`, es decir, la dirección real de la conexión TCP que llegó al servidor Go. Si la API queda detrás de un proxy inverso (como Nginx, Cloudflare o Render), todas las conexiones TCP llegarían con la IP del proxy y el límite de 5 intentos/minuto bloquearía a todos los usuarios a la vez. Sin embargo, no se puede confiar sin más en cabeceras como `X-Forwarded-For` porque cualquier cliente puede escribir o falsificar esa cabecera en su petición y saltarse el límite cambiando el valor en cada intento; solo debe leerse si se valida que la petición proviene de un proxy de confianza.
## Semana 12 · reto: mínimo privilegio en la base
### D5 · Con qué usuario se conecta la API
La API se conecta con el usuario **`api_doublelevel`** (`NOSUPERUSER`, `NOCREATEDB`, `NOCREATEROLE`, sin propiedad sobre ninguna tabla). Sus permisos por tabla son:
- `cuentas`, `paquete_gemas` y `usuarios`: solo `SELECT`, porque en ejecución normal la API solo consulta credenciales para el login y el catálogo de paquetes.
- `pagos` y `misions`: `SELECT, INSERT, UPDATE, DELETE` (más `USAGE, SELECT` sobre `pagos_id_seq` y `misions_id_seq`) porque los endpoints exponen el CRUD completo y cambios de estado.
- `facturas`: `SELECT, INSERT` (más `USAGE, SELECT` sobre `facturas_id_seq`) porque las facturas solo se emiten y consultan, nunca se borran.
Al intentar ejecutar `DROP TABLE pagos` con `api_doublelevel` (`reto_drop.txt`), PostgreSQL respondió `ERROR: must be owner of table pagos`.

### D6 · Quién crea y cambia las tablas
Las tablas las crea y modifica el usuario **`postgres`** (dueño del esquema) únicamente antes del despliegue o cuando cambia algún modelo, ejecutando el comando **`go run . -migrar`** con `DATABASE_URL` apuntando a `postgres`. Si cambiamos un modelo en Go y arrancamos la API normalmente (`go run .`) con `api_doublelevel` sin migrar antes, la API no ejecuta `AutoMigrate` y al consultar o escribir la columna nueva PostgreSQL devuelve un error de columna inexistente en lugar de alterar la estructura.

### D7 · Lo que esto no frena
Si quedara una inyección SQL en la API, un atacante usando `api_doublelevel` **ya no** podría borrar tablas (`DROP TABLE`), alterar esquemas (`ALTER TABLE`), crear superusuarios ni leer otras bases del servidor. Sin embargo, **todavía podría** leer datos sensibles de las tablas donde tiene `SELECT` (como los hashes en `cuentas` o los pagos de otros usuarios) y modificar o borrar filas en `pagos` y `misions`, por lo que las consultas parametrizadas siguen siendo indispensables.
## Semana 13 · los avisos en tiempo real

### D1 · Por qué WebSocket y no SSE ni sondeo
En **DoubleLevel** avisamos al jugador dueño de un `Pago` cuando su estado cambia (`estado_cambiado` con `id` y `estado`), en sentido servidor → cliente. El sondeo no basta porque el aviso puede tardar hasta el intervalo completo de consulta (por ejemplo, 5 o 10 segundos) y genera tráfico innecesario con respuestas vacías. Aunque SSE cubre avisos en un solo sentido, DoubleLevel está pensado como aplicación móvil y web; `EventSource` es propio del navegador, mientras que WebSocket tiene soporte estándar en clientes móviles/escritorio y deja abierta la puerta a comunicación bidireccional futura. El costo de WebSocket es mantener conexiones abiertas en memoria (`canales` en `internal/transporte/avisos.go`), proteger el mapa con `sync.Mutex`, validar el origen manualmente y asumir que no escala entre múltiples instancias sin un broker externo.

### D2 · Quién recibe cada aviso
La llave del mapa `canales` en `internal/transporte/avisos.go` es el ID del usuario convertido a texto (`strconv.Itoa(int(s.UsuarioID))`), y el dueño del registro es el campo `pago.UsuarioID`, que en `crearPago` (`internal/misiones/manejadores.go`) se toma exclusivamente del token mediante `sesion.De(r)` y no del cuerpo JSON. Para que coincidan, tanto `Conectar` (al suscribir la conexión) como `CambiarEstadoPago` en `internal/misiones/servicio.go` (al llamar a `s.avisos.Enviar`) convierten `UsuarioID` con `strconv.Itoa(int(...))`. Si una conexión está lenta y su canal con buffer de 8 se llena, el `select` con `default:` descarta ese aviso para no bloquear al servidor ni a otros usuarios. Con dos copias de la API, cada una tendría su propio mapa en memoria: si `ana` se conecta a la copia A y el `PATCH` de `luis` entra por la copia B, el aviso nunca le llegaría.

### D3 · El token en la dirección
`/ws` va fuera del guardia (`r.Get("/ws", avisos.Conectar)` en `main.go`) porque la API `WebSocket` del navegador no permite enviar la cabecera `Authorization`; si pasara por `tokens.Autenticar`, siempre respondería `401`. El token se valida con `a.tokens.Leer(...)` antes de `websocket.Accept` para rechazar peticiones sin sesión con un `401` HTTP normal sin llegar a abrir el túnel WebSocket. La validación de origen (`OriginPatterns: []string{a.origen}`) sigue siendo necesaria aunque haya token para impedir que una página maliciosa en otro dominio abra conexiones WebSocket contra nuestra API desde el navegador del usuario. Como el token viaja en la URL, quedó registrado en `docs/taller13/logs_avisos.txt` en las líneas `GET /ws?token=...`, por lo que lo reemplazamos por `token=BORRADO` con `sed` antes de subirlo. Si el token vence mientras la conexión ya está abierta, el WebSocket sigue activo porque la firma solo se verifica durante el handshake inicial.