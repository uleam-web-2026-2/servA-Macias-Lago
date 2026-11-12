# Semana 11 · las seis peticiones del hito 2

| Petición | Con go run | En el contenedor | Con Compose |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | 401 |
| 2 login | 200 | 200 | 200 |
| 3 con token | 200 | 200 | 200 |
| 4 rol sin permiso | 403 | 403 | 403 |
| 5 cambio permitido | 200 | 200 | 200 |
| 6 transición prohibida | 422 | 422 | 422 |

Petición 3 con go run (respuesta completa):
```json
[
  {
    "id": 1,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T14:30:00-05:00"
  },
  {
    "id": 2,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T14:30:00-05:00"
  }
]
```

La misma, en el contenedor:
```json
[
  {
    "id": 1,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T19:30:00Z"
  },
  {
    "id": 2,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T19:30:00Z"
  }
]
```

La misma, con Compose:
```json
[
  {
    "id": 1,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T19:30:00Z"
  },
  {
    "id": 2,
    "usuario_id": 1,
    "paquete_gemas_id": 1,
    "monto": 4.99,
    "estado": "pendiente",
    "fecha_creacion": "2026-11-11T19:30:00Z"
  }
]
```