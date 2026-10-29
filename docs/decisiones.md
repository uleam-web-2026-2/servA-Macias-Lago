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

**Estructura elegida:** Arquitectura modular por paquetes y separación estricta de responsabilidades en `internal/misiones/`.

**Qué va en cada archivo y por qué:**
- **`manejadores.go` (Capa HTTP / Transporte):** Recibe las peticiones HTTP, decodifica el JSON de entrada, lee los parámetros de la URL con `chi`, invoca a la base de datos y utiliza el paquete `respuesta` para retornar los códigos HTTP correctos (`200`, `201`, `400`, `401`, `403`, `404`, `422`). **No contiene lógica de negocio compleja.**
- **`estado.go` (Capa de Negocio / Dominio):** Aloja la máquina de estados y la función pura `puedePasar(de, a string) bool`. Esta capa **no sabe absolutamente nada de HTTP** (no importa `net/http` para las reglas de transición) y se limita a validar las reglas del modelo de negocio de los pagos.
- **`modelos.go` (Capa de Persistencia):** Define los structs de GORM (`Pago`, `PaqueteGemas`, `Mision`, etc.) y sus relaciones con la base de datos.