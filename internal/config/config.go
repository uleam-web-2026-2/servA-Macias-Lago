package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	Puerto       string
	TiempoEspera time.Duration
	ClaveFirma   string
}

func Cargar() (Config, error) {
	_ = godotenv.Load()
	var c Config

	c.DatabaseURL = os.Getenv("DATABASE_URL")
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("falta DATABASE_URL (copie .env.example a .env y complete la contraseña)")
	}

	c.ClaveFirma = os.Getenv("CLAVE_FIRMA")
	if c.ClaveFirma == "" {
		return c, fmt.Errorf("falta CLAVE_FIRMA (copie .env.example a .env)")
	}

	c.Puerto = os.Getenv("PUERTO")
	if c.Puerto == "" {
		c.Puerto = "8080"
	}

	segundos := os.Getenv("TIEMPO_ESPERA_SEGUNDOS")
	if segundos == "" {
		segundos = "5"
	}

	n, err := strconv.Atoi(segundos)
	if err != nil || n <= 0 {
		return c, fmt.Errorf("TIEMPO_ESPERA_SEGUNDOS debe ser un entero mayor que cero, llegó %q", segundos)
	}
	c.TiempoEspera = time.Duration(n) * time.Second

	return c, nil
}
