package misiones

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
)

type Manejador struct {
	DB *gorm.DB
}

func (m *Manejador) Rutas(r chi.Router) {
	r.Post("/misiones", m.crear)
	r.Get("/misiones", m.listar)
}

func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	var mision Mision
	if err := json.NewDecoder(r.Body).Decode(&mision); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	mision.ID = 0

	if !estadosValidos[mision.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	if err := m.DB.Debug().Create(&mision).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar la misión")
		return
	}

	respuesta.Exito(w, http.StatusCreated, mision)
}

func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	var lista []Mision
	if err := m.DB.Debug().Find(&lista).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	respuesta.Exito(w, http.StatusOK, lista)
}