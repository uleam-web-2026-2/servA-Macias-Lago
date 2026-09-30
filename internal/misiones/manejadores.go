package misiones

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	r.Get("/misiones/{id}", m.verUno)
	r.Put("/misiones/{id}", m.actualizar)
	r.Delete("/misiones/{id}", m.borrar)
}

func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	var mision Mision
	if err := json.NewDecoder(r.Body).Decode(&mision); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	mision.ID = 0

	// 1. Validar estado (422)
	if !estadosValidos[mision.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	// 2. Título requerido (422)
	if mision.Titulo == "" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "titulo_requerido", "El título de la misión no puede estar vacío")
		return
	}

	// 3. REGLA PROPIA DEL NEGOCIO (Double Level): Dificultad permitida (422)
	if mision.Dificultad != "facil" && mision.Dificultad != "media" && mision.Dificultad != "dificil" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "dificultad_invalida", "La dificultad debe ser facil, media o dificil")
		return
	}

	// *** AQUÍ EMPIEZA LA BASE DE DATOS *** (Línea 53)
	if err := m.DB.Debug().Create(&mision).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar la misión")
		return
	}

	respuesta.Exito(w, http.StatusCreated, mision)
}

func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	var lista []Mision
	estado := r.URL.Query().Get("estado")

	query := m.DB.Debug()

	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if err := query.Find(&lista).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	respuesta.Exito(w, http.StatusOK, lista)
}

func (m *Manejador) verUno(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	var mision Mision
	if err := m.DB.Debug().First(&mision, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	respuesta.Exito(w, http.StatusOK, mision)
}

func (m *Manejador) actualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	var existente Mision
	if err := m.DB.First(&existente, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	var datosNuevos Mision
	if err := json.NewDecoder(r.Body).Decode(&datosNuevos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	if !estadosValidos[datosNuevos.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	if datosNuevos.Titulo == "" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "titulo_requerido", "El título de la misión no puede estar vacío")
		return
	}

	if datosNuevos.Dificultad != "facil" && datosNuevos.Dificultad != "media" && datosNuevos.Dificultad != "dificil" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "dificultad_invalida", "La dificultad debe ser facil, media o dificil")
		return
	}

	existente.Titulo = datosNuevos.Titulo
	existente.Estado = datosNuevos.Estado
	existente.Dificultad = datosNuevos.Dificultad

	if err := m.DB.Debug().Save(&existente).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo actualizar la misión")
		return
	}

	respuesta.Exito(w, http.StatusOK, existente)
}

func (m *Manejador) borrar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	res := m.DB.Debug().Delete(&Mision{}, id)
	if res.Error != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al eliminar de la base de datos")
		return
	}

	if res.RowsAffected == 0 {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Misión no encontrada")
		return
	}

	respuesta.Exito(w, http.StatusOK, map[string]string{"mensaje": "Misión eliminada correctamente"})
}

func leerID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || n <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El id debe ser un número positivo")
		return 0, false
	}
	return uint(n), true
}