package misiones

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/sesion"
)

type Manejador struct {
	Servicio *Servicio
}

func (m *Manejador) Rutas(r chi.Router, tokens *sesion.Tokens) {
	// Nivel 1 Pública: sin token
	r.Get("/paquetes", m.listarPaquetes)
	r.Get("/misiones", m.listar)
	r.Get("/misiones/{id}", m.verUno)

	// Nivel 2 Con sesión: requiere token válido (401 si falta o es inválido)
	conSesion := r.With(tokens.Autenticar)
	conSesion.Post("/misiones", m.crear)
	conSesion.Put("/misiones/{id}", m.actualizar)
	conSesion.Delete("/misiones/{id}", m.borrar)

	conSesion.Get("/pagos", m.listarPagos)
	conSesion.Get("/pagos/mios", m.listarPagos)
	conSesion.Post("/pagos", m.crearPago)
	conSesion.Get("/pagos/{id}", m.verPago)

	// Nivel 3 Solo admin: requiere rol admin (403 si es jugador)
	soloAdmin := conSesion.With(sesion.SoloRol(sesion.RolAdmin))
	soloAdmin.Put("/pagos/{id}", m.actualizarPago)
	soloAdmin.Delete("/pagos/{id}", m.borrarPago)
	soloAdmin.Patch("/pagos/{id}", m.cambiarEstado)
	soloAdmin.Patch("/pagos/{id}/estado", m.cambiarEstado)
}

func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var pedido struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "el cuerpo no es JSON válido")
		return
	}

	pago, err := m.Servicio.CambiarEstadoPago(id, pedido.Estado)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no existe el pago")
			return
		}
		respuesta.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	respuesta.JSON(w, http.StatusOK, pago)
}

func (m *Manejador) listarPaquetes(w http.ResponseWriter, r *http.Request) {
	paquetes, err := m.Servicio.ListarPaquetes()
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar los paquetes de gemas")
		return
	}
	respuesta.Exito(w, http.StatusOK, paquetes)
}

func (m *Manejador) crearPago(w http.ResponseWriter, r *http.Request) {
	var pago Pago
	if err := json.NewDecoder(r.Body).Decode(&pago); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}

	creado, err := m.Servicio.CrearPago(pago)
	if err != nil {
		respuesta.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	respuesta.Exito(w, http.StatusCreated, creado)
}

func (m *Manejador) listarPagos(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	pagos, err := m.Servicio.ListarPagos(estado)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar los pagos")
		return
	}
	respuesta.Exito(w, http.StatusOK, pagos)
}

func (m *Manejador) verPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	pago, err := m.Servicio.BuscarPago(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Pago no encontrado")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	respuesta.Exito(w, http.StatusOK, pago)
}

func (m *Manejador) actualizarPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	existente, err := m.Servicio.BuscarPago(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Pago no encontrado")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}

	var datosNuevos Pago
	if err := json.NewDecoder(r.Body).Decode(&datosNuevos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}
	if datosNuevos.PaqueteGemasID == 0 || datosNuevos.Monto <= 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "paquete_gemas_id y monto son obligatorios")
		return
	}

	existente.PaqueteGemasID = datosNuevos.PaqueteGemasID
	existente.Monto = datosNuevos.Monto
	existente.ComprobanteURL = datosNuevos.ComprobanteURL

	if err := m.Servicio.GuardarPago(&existente); err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "No se pudo actualizar el pago")
		return
	}
	respuesta.Exito(w, http.StatusOK, existente)
}

func (m *Manejador) borrarPago(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	filas, err := m.Servicio.BorrarPago(id)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al eliminar el pago")
		return
	}
	if filas == 0 {
		respuesta.Error(w, http.StatusNotFound, "Pago no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Manejadores de Misiones
func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	var pedido MisionPeticion
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}
	creada, err := m.Servicio.CrearMision(pedido)
	if err != nil {
		respuesta.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	respuesta.Exito(w, http.StatusCreated, NuevaMisionRespuesta(creada))
}

func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	lista, err := m.Servicio.ListarMisiones(estado)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	// Mapear lista a DTOs de respuesta
	var respuestas []MisionRespuesta
	for _, mis := range lista {
		respuestas = append(respuestas, NuevaMisionRespuesta(mis))
	}
	respuesta.Exito(w, http.StatusOK, respuestas)
}

func (m *Manejador) verUno(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	mision, err := m.Servicio.BuscarMision(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	respuesta.Exito(w, http.StatusOK, NuevaMisionRespuesta(mision))
}

func (m *Manejador) actualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var pedido MisionPeticion
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}
	actualizada, err := m.Servicio.ActualizarMision(id, pedido)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	respuesta.Exito(w, http.StatusOK, NuevaMisionRespuesta(actualizada))
}

func (m *Manejador) borrar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	filas, err := m.Servicio.BorrarMision(id)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al eliminar de la base de datos")
		return
	}
	if filas == 0 {
		respuesta.Error(w, http.StatusNotFound, "Misión no encontrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func leerID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || n <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "El id debe ser un número positivo")
		return 0, false
	}
	return uint(n), true
}