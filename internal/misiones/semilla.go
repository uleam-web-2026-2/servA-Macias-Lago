package misiones

import "gorm.io/gorm"

func Sembrar(db *gorm.DB) {
	var totalUsuarios int64
	db.Model(&Usuario{}).Count(&totalUsuarios)
	if totalUsuarios == 0 {
		u := Usuario{
			Nombre: "Gia",
			Correo: "gia@doublelevel.ec",
			Rol:    "jugador",
			Clase:  "Aprendiz",
			Nivel:  1,
			Misiones: []Mision{
				{Titulo: "Estudiar Go y GORM", Categoria: "Estudio", Dificultad: "media", Estado: "pendiente"},
			},
		}
		db.Create(&u)
	}

	var totalPaquetes int64
	db.Model(&PaqueteGemas{}).Count(&totalPaquetes)
	if totalPaquetes == 0 {
		paquetes := []PaqueteGemas{
			{Nombre: "Cofre de 500 Gemas", Precio: 4.99, CantidadGemas: 500},
			{Nombre: "Cofre Real de 1200 Gemas", Precio: 9.99, CantidadGemas: 1200},
		}
		db.Create(&paquetes)
	}

	var totalPagos int64
	db.Model(&Pago{}).Count(&totalPagos)
	if totalPagos == 0 {
		pagos := []Pago{
			{UsuarioID: 1, PaqueteGemasID: 1, Monto: 4.99, Estado: "pendiente"},
			{UsuarioID: 1, PaqueteGemasID: 1, Monto: 4.99, Estado: "pendiente"},
		}
		db.Create(&pagos)
	}
}
