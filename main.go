package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/middleware"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/tickets"
)

func main() {
	// Cadena con sslmode=disable para evitar errores TLS locales
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5433/mesa_ayuda?sslmode=disable")
	if err != nil {
		log.Fatalf("Error al abrir conexion con la BD: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Registro)
	r.Use(middleware.Recuperacion)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusNotFound, "ruta_inexistente", "la ruta no existe")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusMethodNotAllowed, "metodo_no_permitido", "el metodo no esta permitido en esta ruta")
	})

	r.Get("/salud", func(w http.ResponseWriter, r *http.Request) {
		var version string
		if err := db.QueryRow("select version()").Scan(&version); err != nil {
			respuesta.Error(w, http.StatusInternalServerError, "error_bd", fmt.Sprintf("sin conexion: %v", err))
			return
		}
		respuesta.Exito(w, http.StatusOK, map[string]string{"bd": "ok", "version": version})
	})

	almacen := tickets.NuevoAlmacen()
	r.Get("/tickets", almacen.Listar)
	r.Post("/tickets", almacen.Crear)
	r.Get("/tickets/{id}", almacen.Obtener)
	r.Get("/explotar", tickets.Explotar)

	log.Println("Servidor Chi con PostgreSQL escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}