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
	// Rutas de Misiones
	r.Post("/misiones", m.crear)
	r.Get("/misiones", m.listar)
	r.Get("/misiones/{id}", m.verUno)
	r.Put("/misiones/{id}", m.actualizar)
	r.Delete("/misiones/{id}", m.borrar)

	// Rutas de Monetización (Sección 7 de la Ficha del Negocio)
	r.Get("/paquetes", m.listarPaquetes)
	r.Get("/pagos", m.listarPagos)
	r.Get("/pagos/mios", m.listarPagos)
	r.Post("/pagos", m.crearPago)
	r.Get("/pagos/{id}", m.verPago)
	r.Put("/pagos/{id}", m.actualizarPago)
	r.Delete("/pagos/{id}", m.borrarPago)
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
	existente.Dificultad = datosNuevos.Dificultad // El estado solo cambia con PATCH

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

// ==========================================
// MANEJADORES DE PAGOS Y PAQUETES (FICHA)
// ==========================================

func (m *Manejador) listarPaquetes(w http.ResponseWriter, r *http.Request) {
	var paquetes []PaqueteGemas
	if err := m.DB.Debug().Find(&paquetes).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar los paquetes de gemas")
		return
	}
	respuesta.Exito(w, http.StatusOK, paquetes)
}

func (m *Manejador) crearPago(w http.ResponseWriter, r *http.Request) {
	var pago Pago
	if err := json.NewDecoder(r.Body).Decode(&pago); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	pago.ID = 0
	if pago.Estado == "" {
		pago.Estado = "pendiente"
	}

	if !estadosPagoValidos[pago.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado de pago no permitido")
		return
	}

	if pago.PaqueteGemasID == 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "paquete_requerido", "El paquete_gemas_id es obligatorio")
		return
	}

	if pago.Monto <= 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "monto_invalido", "El monto debe ser mayor a cero")
		return
	}

	if err := m.DB.Debug().Create(&pago).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo registrar el pago")
		return
	}

	respuesta.Exito(w, http.StatusCreated, pago)
}

func (m *Manejador) listarPagos(w http.ResponseWriter, r *http.Request) {
	var pagos []Pago
	estado := r.URL.Query().Get("estado")

	query := m.DB.Debug()
	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if err := query.Find(&pagos).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar los pagos")
		return
	}

	respuesta.Exito(w, http.StatusOK, pagos)
}

func (m *Manejador) verPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	var pago Pago
	if err := m.DB.Debug().First(&pago, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Pago no encontrado")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	respuesta.Exito(w, http.StatusOK, pago)
}

func (m *Manejador) actualizarPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	var existente Pago
	if err := m.DB.First(&existente, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Pago no encontrado")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al consultar la base de datos")
		return
	}

	var datosNuevos Pago
	if err := json.NewDecoder(r.Body).Decode(&datosNuevos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	if datosNuevos.PaqueteGemasID == 0 || datosNuevos.Monto <= 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "datos_invalidos", "paquete_gemas_id y monto son obligatorios")
		return
	}

	existente.PaqueteGemasID = datosNuevos.PaqueteGemasID
	existente.Monto = datosNuevos.Monto
	existente.ComprobanteURL = datosNuevos.ComprobanteURL
	// El estado solo se cambia con PATCH

	if err := m.DB.Debug().Save(&existente).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo actualizar el pago")
		return
	}

	respuesta.Exito(w, http.StatusOK, existente)
}

func (m *Manejador) borrarPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	res := m.DB.Debug().Delete(&Pago{}, id)
	if res.Error != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al eliminar el pago")
		return
	}

	if res.RowsAffected == 0 {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Pago no encontrado")
		return
	}

	respuesta.Exito(w, http.StatusOK, map[string]string{"mensaje": "Pago eliminado correctamente"})
}

func leerID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || n <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El id debe ser un número positivo")
		return 0, false
	}
	return uint(n), true
}
