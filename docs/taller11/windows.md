# Notas de ejecución en Windows

##Conversión de rutas POSIX en Git Bash (MINGW64)
- **Qué falló:** Al ejecutar `docker run --rm doublelevel:compilacion ls -a /src`, el contenedor devolvió `ls: cannot access 'C:/Program Files/Git/src': No such file or directory`.
- **Por qué:** La terminal Git Bash (MSYS2/MINGW64) en Windows convierte automáticamente los argumentos que inician con `/` a rutas locales de Windows antes de pasarlos al binario de Docker.
- **Qué hicimos:** Usamos doble barra (`//src`) en el comando (`docker run --rm doublelevel:compilacion ls -a //src > docs/taller11/contexto.txt`) para desactivar la conversión automática de rutas de MINGW64.