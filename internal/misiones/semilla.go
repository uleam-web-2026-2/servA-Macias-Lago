package misiones

import "gorm.io/gorm"

func Sembrar(db *gorm.DB) {
	var total int64
	db.Model(&Usuario{}).Count(&total)
	if total > 0 {
		return
	}

	usuarios := []Usuario{
		{
			Nombre: "Gia",
			Clase:  "Aprendiz",
			Nivel:  1,
			Misiones: []Mision{
				{Titulo: "Estudiar Go y GORM", Categoria: "Estudio", Dificultad: "Media", Estado: "pendiente"},
			},
		},
	}

	db.Create(&usuarios)
}