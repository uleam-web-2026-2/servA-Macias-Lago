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