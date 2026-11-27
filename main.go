package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/config"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/sesion"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/transporte"
)

// origenFront es el único front que puede leer las respuestas (CORS) y
// abrir el WebSocket. Se escribe una vez, igual que en la barra del navegador.
const origenFront = "http://localhost:5173"

func main() {
	migrar := flag.Bool("migrar", false, "ejecuta migraciones y semillas")
	flag.Parse()

	cfg, err := config.Cargar()
	if err != nil {
		log.Fatal("configuración: ", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("no se pudo conectar: ", err)
	}

	if *migrar {
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
		log.Println("migraciones y semillas completadas")
		return
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{origenFront},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))
	r.Use(middleware.ClientIPFromRemoteAddr) // anota la IP del cliente
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tokens := sesion.NuevosTokens(cfg.ClaveFirma)
	login := &sesion.Manejador{DB: db, Tokens: tokens}
	// Pública, pero con límite: 5 intentos por minuto por IP. El sexto, 429.
	r.With(httprate.LimitBy(5, time.Minute, porIP)).Post("/login", login.Login)

	conexion, err := db.DB()
	if err != nil {
		log.Fatal("no se pudo obtener la conexión: ", err)
	}

	repoGorm := misiones.NuevoAlmacenGorm(db)
	repo := misiones.NuevoMisionesSqlc(repoGorm, conexion)
	servicio := misiones.NuevoServicio(repo)

	// Los avisos: el servicio avisa al dueño cuando cambia su pago.
	avisos := transporte.NuevosAvisos(tokens, origenFront)
	servicio.AvisarCon(avisos)

	(&misiones.Manejador{Servicio: servicio}).Rutas(r, tokens)
	r.Get("/ws", avisos.Conectar) // fuera del guardia: el token va en la dirección

	servidor := &http.Server{
		Addr:         ":" + cfg.Puerto,
		Handler:      r,
		ReadTimeout:  cfg.TiempoEspera,
		WriteTimeout: cfg.TiempoEspera,
	}

	log.Println("escuchando en el puerto", cfg.Puerto)
	log.Fatal(servidor.ListenAndServe())
}

// porIP dice a quién se le cuenta cada petición: a la IP que anotó
// ClientIPFromRemoteAddr, la de la conexión. No se lee ninguna cabecera
// que el cliente pueda escribir.
func porIP(r *http.Request) (string, error) {
	ip := middleware.GetClientIP(r.Context())
	return httprate.CanonicalizeIP(ip), nil
}