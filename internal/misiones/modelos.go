package misiones

import "gorm.io/gorm"

// Usuario representa la entidad del lado del UNO
type Usuario struct {
	gorm.Model
	Nombre   string   `json:"Nombre"`
	Clase    string   `json:"Clase"`
	Nivel    int      `json:"Nivel"`
	Misiones []Mision `json:"Misiones"` // Relación 1 a N
}

// Mision representa la entidad con estados del lado de los MUCHOS
type Mision struct {
	gorm.Model
	Titulo     string `json:"Titulo"`
	Categoria  string `json:"Categoria"`
	Dificultad string `json:"Dificultad"`
	Estado     string `json:"Estado"`
	UsuarioID  uint   `json:"UsuarioID"` // Clave foránea
}

var estadosValidos = map[string]bool{
	"pendiente":  true,
	"completada": true,
	"fallida":    true,
}