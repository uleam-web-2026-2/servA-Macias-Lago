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

## Semana 11 — el contenedor

### D1 · Qué entra en la imagen
La imagen final (`doublelevel:latest`, que pesa **40.6 MB** en disco y **12.3 MB** de contenido según `imagenes.txt`) contiene únicamente la base mínima `gcr.io/distroless/static-debian13:nonroot` (certificados raíz TLS, `tzdata` y archivos `/etc/passwd` para el usuario `65532`) más nuestro binario estático `/servidor` (**22.5 MB** según `capas.txt`), compilado con `CGO_ENABLED=0`. En la primera etapa (`doublelevel:compilacion`, de **1.64 GB** / **392 MB**) se quedaron el compilador de Go, las herramientas del sistema Debian, el código fuente (`.go`) y la caché de módulos descargados.
Si no existiera `.dockerignore`, el archivo `.env` **tampoco llegaría a la imagen final** porque la segunda etapa solo ejecuta `COPY --from=compilacion /servidor /servidor`; sin embargo, **sí llegaría a la primera etapa (`compilacion`)** dentro de `/src` al hacer `COPY . .`, quedando expuesto en las capas intermedias de construcción de Docker.

### D2 · Dónde está la base, vista desde el contenedor
Falló `localhost:5433` (`dial tcp 127.0.0.1:5433: connect: connection refused`) porque cada contenedor tiene su propio aislamiento de red (network namespace): dentro del contenedor `doublelevel-api`, `localhost` (`127.0.0.1`) apunta al propio contenedor de la API y no a nuestra laptop Windows, y dentro de ese contenedor no hay ningún PostgreSQL escuchando. En su lugar pusimos `host.docker.internal:5433`, que Docker Desktop resuelve a la IP de la máquina anfitriona donde está publicado el contenedor `pg`. Cuando la API corra en un servidor real en la nube (o en una red de contenedores), `host.docker.internal` deberá cambiar por el nombre DNS del servicio de base de datos en la red interna o el endpoint gestionado de producción.

### D3 · Los puertos
Usamos `-p 8080:8080`. El número de la izquierda (`8080`) es el puerto publicado en nuestra máquina anfitriona (al que llama Bruno en `http://localhost:8080`), y el número de la derecha (`8080`) es el puerto interno del contenedor donde escucha el proceso `/servidor`. El número que **tiene que coincidir obligatoriamente con la variable `PUERTO`** de `.env.docker` es el de la **derecha** (el del contenedor). Si no coincide, Docker redirige el tráfico al puerto equivocado del contenedor donde nadie escucha y el cliente recibe fallo de conexión.

### D4 · La hora
En el paso 2.5 comprobamos que **sí es exactamente el mismo instante en el tiempo** (`2026-11-11T14:30:00-05:00` en Ecuador equivale a `2026-11-11T19:30:00Z` en UTC, con las 5 horas de diferencia de `America/Guayaquil`). Cambió la representación porque con `go run` el binario lee la zona horaria local configurada en Windows (`-05:00`), mientras que el contenedor `distroless` no tiene configurada una zona local por defecto y funciona en tiempo universal coordinado (**UTC**, sufijo `Z`), que además es cómo PostgreSQL entrega los campos `TIMESTAMPTZ` (`fecha_creacion`). Decidimos mantener el servidor en UTC (`Z`) dentro del contenedor porque es el estándar ISO-8601 en APIs REST y evita ambigüedades entre clientes de distintas zonas horarias.


## Semana 11 — reto: la API y la base, juntas

### D5 · Cómo encuentra la API a la base
La API se conecta usando la dirección `db` y el puerto `5432` (`postgres://postgres:${POSTGRES_PASSWORD}@db:5432/doublelevel_db?sslmode=disable`). Usa `db` porque Docker Compose crea una red interna propia y registra el nombre de cada servicio como un nombre DNS interno. Usa el puerto `5432` y no el `5433` porque la comunicación ocurre directamente de contenedor a contenedor dentro de la red virtual de Compose, donde PostgreSQL escucha en su puerto nativo (`5432`); el `5433` era únicamente un desvío hacia la máquina anfitriona Windows para el contenedor suelto de la semana 3.

### D6 · El orden de arranque
La primera vez, con la base recién creada y vacía, PostgreSQL tarda unos segundos en inicializar el clúster (`/var/lib/postgresql/data`) y crear `doublelevel_db`. Si la API arranca al mismo instante, intenta conectarse antes de que PostgreSQL acepte conexiones TCP y el proceso de Go termina con `log.Fatal("no se pudo conectar")`. Lo resolvimos agregando un `healthcheck` en el servicio `db` (`pg_isready -U postgres -d doublelevel_db`) y configurando en `api` la condición `depends_on: db: condition: service_healthy`, de modo que Docker espera a que la base esté lista para recibir consultas antes de iniciar `/servidor`.

### D7 · Dónde quedan los datos
Los registros sobreviven a `docker compose down` porque declaramos un volumen nombrado administrado por Docker (`datos_pg:/var/lib/postgresql/data`), el cual persiste en el disco del motor de Docker fuera del ciclo de vida del contenedor. Lo único que borra esos datos para probar desde cero es ejecutar `docker compose down -v` (la bandera `-v` elimina explícitamente los volúmenes asociados al proyecto).

### D8 · Los secretos
Si `compose.yaml` no tiene contraseñas ni claves escritas en texto plano, los valores salen de la interpolación de variables (`${POSTGRES_PASSWORD}`, `${CLAVE_FIRMA}`) que Docker Compose lee automáticamente del archivo `.env` ubicado en la misma carpeta raíz cuando se ejecuta `docker compose up` (y que está excluido en `.gitignore`).

### D9 · Lo que se publica en la máquina
Nuestro `compose.yaml` publica en la máquina anfitriona únicamente el puerto del servicio `api` (`${PUERTO}:${PUERTO}`, es decir, `8080:8080`), porque es el único punto de entrada que necesita consumir Bruno desde fuera. El servicio `db` **no** tiene sección `ports:`, lo que evita cualquier choque con el contenedor `pg` de la semana 3 y protege la base de accesos externos. Cuando necesitemos consultarla con `psql`, entramos directamente al contenedor con: `docker compose exec db psql -U postgres -d doublelevel_db`.