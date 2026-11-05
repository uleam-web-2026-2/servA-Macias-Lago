package misiones

import "time"

// DTO para la petición de creación/actualización de Misión
type MisionPeticion struct {
	Titulo     string `json:"titulo"`
	Categoria  string `json:"categoria"`
	Dificultad string `json:"dificultad"`
	Estado     string `json:"estado"`
}

func (p MisionPeticion) AEntidad() Mision {
	return Mision{
		Titulo:     p.Titulo,
		Categoria:  p.Categoria,
		Dificultad: p.Dificultad,
		Estado:     p.Estado,
	}
}

// DTO para la respuesta de Misión
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