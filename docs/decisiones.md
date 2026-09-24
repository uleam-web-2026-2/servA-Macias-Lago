# Decisiones

## D1 · ¿Quién pone el estado al crear?
**Opciones:** que lo mande quien crea, o que lo ponga el servidor.
**Qué elegimos:** Que el estado lo envíe quien crea el registro en la petición JSON.
**Por qué, en nuestro negocio:** Porque las misiones pueden crearse directamente en distintos estados según la planificación inicial.
**Qué pasaría con la otra opción:** Si el servidor fijara siempre un estado por defecto, obligaría a realizar un `PUT` adicional de inmediato para ajustar aquellas misiones que inician directamente en proceso.