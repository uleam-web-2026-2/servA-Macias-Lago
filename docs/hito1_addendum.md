# Hito 1 · Addendum técnico

**Pareja:** Lago · Macias (Pareja 15)  
**Paralelo:** Aplicaciones Web II A  

---

## A. Estructura del proyecto

```text
.
├── main.go                       # Punto de entrada de la aplicación Go
├── go.mod                        # Definición de módulo y dependencias
├── go.sum                        # Sumas de verificación de dependencias
├── internal/
│   └── misiones/             # Paquete principal con la lógica de negocio
│       ├── modelos.go            # Declaración de structs GORM y JSON
│       ├── manejadores.go        # Lógica de endpoints y controladores HTTP
│       └── rutas.go              # Mapeo de rutas y registro de middlewares
└── docs/                         # Documentación del proyecto y recursos multimedia
    ├── hito1_ficha_del_negocio.md
    ├── hito1_addendum.md
    ├── hito1_modelo.png
    ├── hito1_estados.png
    ├── hito1_pruebas.png
    ├── hito1_boceto.png
    ├── hito1_respuesta_ok.png
    └── hito1_respuesta_error.png
```

---

## B. Configuración y secretos

| Variable | Para qué sirve | Ejemplo (sin datos reales) |
|----------|----------------|----------------------------|
| `PORT` | Puerto HTTP donde se ejecuta el servidor Go | `8080` |
| `DATABASE_URL` | Cadena de conexión a la base de datos PostgreSQL/SQLite | `postgres://user:pass@localhost:5432/doublelevel_db?sslmode=disable` |
| `JWT_SECRET` | Clave secreta para la firma de tokens de autenticación | `clave_secreta_super_segura_ejemplo` |

**Archivo de ejemplo:** `.env.example`

---
## C. Pruebas

| Prueba | Qué caso cubre |
|--------|----------------|
| `TestCrear` | Valida la creación enviando datos correctos en el cuerpo de la petición. |
| `TestCrear/JSON_roto_responde_400` | Verifica que enviar un JSON mal estructurado devuelva un código HTTP 400 Bad Request. |
| `TestCrear/titulo_vacio_responde_422` | Comprueba que si falta el título en la petición devuelva un código HTTP 422 Unprocessable Entity. |
| `TestCrear/estado_invalido_responde_422` | Valida que enviar un estado no permitido en la máquina de estados devuelva un código HTTP 422. |
| `TestCrear/dificultad_no_permitida_responde_422` | Valida las reglas de negocio propias del backend respondiendo con HTTP 422 si los parámetros no están autorizados. |
| `TestVerUnoConIDQueNoEsNumeroResponde400` | Comprueba que al consultar una ruta con un ID no numérico (por ejemplo `/misiones/abc`) devuelva un código HTTP 400. |

**Captura de `go test ./...`:** ![Pruebas](hito1_pruebas.png)

---

## D. Boceto de l
a pantalla principal

La pantalla "Historial de Compras" muestra las compras de gemas realizadas por el usuario, indicando el paquete adquirido, el monto, la fecha y la etiqueta formateada con el estado del pago (`pendiente`, `pago_por_verificar`, `completado`, `rechazado`).

![Boceto](hito1_boceto.png)

---

## E. Diagrama de secuencia del caso de uso principal

**Caso de uso:** Jugador registra transferencia bancaria y sube comprobante.

```mermaid
sequenceDiagram
    autonumber
    actor Jugador
    participant App as App Móvil (Tienda)
    participant API as API Servidor Go
    participant BD as Base de Datos

    Jugador->>App: Selecciona "Cargar comprobante de pago"
    App->>API: POST /pagos (paquete_id, monto, comprobante_url)
    API->>BD: Inserta registro de Pago (estado='pago_por_verificar')
    BD-->>API: Confirmación de registro guardado
    API-->>App: 201 Created (Objeto Pago actualizado)
    App-->>Jugador: Muestra "Pago en verificación por un administrador"
```

---

## F. Capturas de respuestas

**Caso correcto (`GET /paquetes` - 200 OK):**  
![Respuesta correcta](hito1_respuesta_ok.png)

**Caso con error de validación (`POST /pagos` sin `paquete_id` - 422 Unprocessable Entity):**  
![Respuesta con error](hito1_respuesta_error.png)