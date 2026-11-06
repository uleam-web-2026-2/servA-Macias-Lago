package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/config"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/sesion"
)

func main() {
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatal("configuración: ", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("no se pudo conectar: ", err)
	}

	if err := db.AutoMigrate(
		&misiones.Usuario{},
		&misiones.Mision{},
		&misiones.PaqueteGemas{},
		&misiones.Pago{},
		&misiones.Factura{},
		&misiones.ConsumoGemas{},
		&sesion.Usuario{},
	); err != nil {
		log.Fatal("no se pudo migrar: ", err)
	}

	misiones.Sembrar(db)
	if err := sesion.Sembrar(db); err != nil {
		log.Fatal("no se pudieron sembrar los usuarios de sesión: ", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tokens := sesion.NuevosTokens(cfg.ClaveFirma)
	login := &sesion.Manejador{DB: db, Tokens: tokens}
	r.Post("/login", login.Login)

	conexion, err := db.DB()
	if err != nil {
		log.Fatal("no se pudo obtener la conexión: ", err)
	}

	repoGorm := misiones.NuevoAlmacenGorm(db)
	repo := misiones.NuevoMisionesSqlc(repoGorm, conexion)
	servicio := misiones.NuevoServicio(repo)
	(&misiones.Manejador{Servicio: servicio}).Rutas(r, tokens)

	servidor := &http.Server{
		Addr:         ":" + cfg.Puerto,
		Handler:      r,
		ReadTimeout:  cfg.TiempoEspera,
		WriteTimeout: cfg.TiempoEspera,
	}

	log.Println("escuchando en el puerto", cfg.Puerto)
	log.Fatal(servidor.ListenAndServe())
}
