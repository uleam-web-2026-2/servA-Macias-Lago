package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/middleware"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
)

func main() {
	reset := flag.Bool("reset", false, "borra las tablas y arranca con la base vacia")
	flag.Parse()

	// Conexión al puerto 5433 y a la BD Doublevel_db creada en Docker
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=doublevel_db sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error al conectar a la BD: %v", err)
	}

	if *reset {
		db.Migrator().DropTable(&misiones.Mision{}, &misiones.Usuario{})
	}

	err = db.Debug().AutoMigrate(&misiones.Usuario{}, &misiones.Mision{})
	if err != nil {
		log.Fatalf("Error en AutoMigrate: %v", err)
	}

	misiones.Sembrar(db)

	r := chi.NewRouter()
	r.Use(middleware.Registro)
	r.Use(middleware.Recuperacion)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusNotFound, "ruta_inexistente", "la ruta no existe")
	})

	(&misiones.Manejador{DB: db}).Rutas(r)

	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}