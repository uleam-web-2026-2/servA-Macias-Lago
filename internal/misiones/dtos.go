package misiones

import "time"

// --- DTOs para Pago (Entidad con máquina de estados) ---

// PagoPeticion no acepta id, ni estado, ni usuario_id del cuerpo
type PagoPeticion struct {
	PaqueteGemasID uint    `json:"paquete_gemas_id"`
	Monto          float64 `json:"monto"`
	ComprobanteURL *string `json:"comprobante_url,omitempty"`
}

func (p PagoPeticion) AEntidad() Pago {
	return Pago{
		PaqueteGemasID: p.PaqueteGemasID,
		Monto:          p.Monto,
		ComprobanteURL: p.ComprobanteURL,
	}
}

type PagoRespuesta struct {
	ID             uint      `json:"id"`
	UsuarioID      uint      `json:"usuario_id"`
	PaqueteGemasID uint      `json:"paquete_gemas_id"`
	Monto          float64   `json:"monto"`
	Estado         string    `json:"estado"`
	ComprobanteURL *string   `json:"comprobante_url,omitempty"`
	FechaCreacion  time.Time `json:"fecha_creacion"`
}

func NuevoPagoRespuesta(p Pago) PagoRespuesta {
	return PagoRespuesta{
		ID:             p.ID,
		UsuarioID:      p.UsuarioID,
		PaqueteGemasID: p.PaqueteGemasID,
		Monto:          p.Monto,
		Estado:         p.Estado,
		ComprobanteURL: p.ComprobanteURL,
		FechaCreacion:  p.FechaCreacion,
	}
}

// --- DTOs para Misión ---

type MisionPeticion struct {
	Titulo     string `json:"titulo"`
	Categoria  string `json:"categoria"`
	Dificultad string `json:"dificultad"`
}

func (p MisionPeticion) AEntidad() Mision {
	return Mision{
		Titulo:     p.Titulo,
		Categoria:  p.Categoria,
		Dificultad: p.Dificultad,
	}
}

type MisionRespuesta struct {
	ID         uint      `json:"id"`
	Titulo     string    `json:"titulo"`
	Categoria  string    `json:"categoria"`
	Dificultad string    `json:"dificultad"`
	Estado     string    `json:"estado"`
	CreatedAt  time.Time `json:"created_at"`
}

func NuevaMisionRespuesta(m Mision) MisionRespuesta {
	return MisionRespuesta{
		ID:         m.ID,
		Titulo:     m.Titulo,
		Categoria:  m.Categoria,
		Dificultad: m.Dificultad,
		Estado:     m.Estado,
		CreatedAt:  m.CreatedAt,
	}
}
