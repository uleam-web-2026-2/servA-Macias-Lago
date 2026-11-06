# Decisiones

## D1 · ¿Quién pone el estado al crear?
**Opciones:** que lo mande quien crea, o que lo ponga el servidor.
**Qué elegimos:** Que el estado lo envíe quien crea el registro en la petición JSON.
**Por qué, en nuestro negocio:** Porque las misiones pueden crearse directamente en distintos estados según la planificación inicial.
**Qué pasaría con la otra opción:** Si el servidor fijara siempre un estado por defecto, obligaría a realizar un `PUT` adicional de inmediato para ajustar aquellas misiones que inician directamente en proceso.

# Registro de Decisiones de Arquitectura

## Semana 4

### D1 · Si falta una variable al arrancar
**Decisión:** `DATABASE_URL` es una variable obligatoria sin valor por defecto que detiene la ejecución si no está presente. `PUERTO` y `TIEMPO_ESPERA_SEGUNDOS` toman `8080` y `5` respectivamente.
**Por qué:** En *Double Level*, la cadena de conexión contiene las credenciales privadas de PostgreSQL que cambian según el entorno. El puerto y timeout no contienen secretos y facilitan el arranque local rápido.              
**Qué descartamos:** Mantener credenciales en `main.go` o predefinir una contraseña por defecto en el código fuente.

### D2 · Qué hicimos con la contraseña que ya está en el historial
**Decisión:** Se ejecutó un cambio de contraseña en PostgreSQL (`ALTER USER`) y se aisló el valor sensible en el archivo `.env` mediante la variable `DATABASE_URL`.
**Por qué:** Cualquier credencial presente en el historial de Git está expuesta. Cambiar la clave activa invalida los valores presentes en los commits anteriores.
**Qué descartamos:** Reescribir el historial de Git sin modificar la clave real en la base de datos.

### D3 · La prueba que no escribimos
**Prueba que no escribimos:** Creación exitosa de una misión que responda HTTP 201 Created.
**Dónde empieza la base:** En `internal/misiones/manejadores.go`, en la línea **53** (donde se ejecuta `m.DB.Debug().Create(&mision)`).
**Qué haría falta para probarla:** Se requiere un mock de la interfaz de base de datos o levantar una instancia de PostgreSQL en un contenedor de prueba (Integration Testing) para evitar un pánico por puntero nulo (`nil pointer dereference`).

## Semana 8 Estructura · Arquitectura en Capas

**Estructura elegida:**Separación por responsabilidades dentro del paquete de dominio `internal/misiones/` en tres capas: Transporte HTTP, Reglas del Negocio y Repositorio de Datos.

**Qué va en cada archivo y por qué:**
- **`manejadores.go` (Capa HTTP / Transporte):** Recibe la petición HTTP, valida el formato JSON y parámetros de ruta, invoca a la regla del negocio y al repositorio, y traduce el resultado a códigos HTTP (`200`, `201`, `400`, `401`, `403`, `404`, `422`). **No consulta la base de datos directamente.**
- **`estado.go` (Capa de Negocio / Dominio):** Contiene la máquina de estados (`siguientes`), `puedePasar` y `ValidarCambioEstado`. **No importa `net/http` ni `gorm`**, manteniendo las reglas del negocio aisladas del transporte y la base de datos.
- **`repositorio.go` y `modelos.go` (Capa de Persistencia / Datos):** Encapsula todas las consultas de GORM (`BuscarPago`, `GuardarPago`, `CrearPago`, `ListarPagos`) hacia PostgreSQL.

## Semana 10 — transacciones

### D1 · El caso
En la operación **`CompletarMisionYOtorgarRecompensa(misionID, usuarioID)`**, el servicio ejecuta dos escrituras atómicas: cambia el estado de la `Mision` a `"completada"` (tabla `misions`) y actualiza la experiencia/gemas del `Usuario` (tabla `usuarios`). Si el segundo paso falla, la misión quedaría cerrada sin otorgar la recompensa.

### D2 · Opciones que consideramos
1. **Opción A (Manejar transacciones dentro del Servicio):** Rompe la regla de dependencia y las pruebas unitarias sin base de datos.
2. **Opción B (Encapsular la transacción dentro del adaptador del Repositorio):** El servicio delega la acción y cada adaptador (`AlmacenGorm` o `MisionesSqlc`) gestiona su propia transacción internamente.

### D3 · La que elegimos y por qué
Elegimos la **Opción B** porque aísla la infraestructura del dominio, permitiendo pruebas en memoria con repositorios falsos y cumpliendo todas las restricciones.

### D4 · La que descartamos y por qué
Descartamos la Opción A por acoplar el servicio a motores de base de datos concretos, y descartamos la compensación manual por riesgo de inconsistencia ante fallos de red.

### D5 · Con sqlc
Utilizamos el método `WithTx(tx *sql.Tx) *Queries` generado por `sqlc`. El adaptador abre `a.Conexion.Begin()`, envuelve las consultas con `qtx := a.q.WithTx(tx)` y aplica `Commit()` o `Rollback()` según corresponda.

### D6 · Archivos
- **Archivos tocados:** `sqlc.yaml`, `internal/misiones/schema.sql`, `internal/misiones/query.sql`, `internal/misiones/sqlc/*`, `internal/misiones/dtos.go`, `internal/misiones/repositorio.go`, `internal/misiones/servicio.go`, `internal/misiones/manejadores.go`, `main.go`, `docs/openapi.yaml`, `.env.example`, `docs/taller10/*` y `docs/decisiones.md`.
- **Archivos sin cambios:** `internal/config/*`, `internal/sesion/*`, `internal/respuesta/*`.