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
	Repo *Repositorio
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
	if _, existe := siguientes[pedido.Estado]; !existe {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado no permitido: "+pedido.Estado)
		return
	}

	pago, err := m.Repo.BuscarPago(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respuesta.Error(w, http.StatusNotFound, "no existe el pago")
		return
	}
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "no se pudo leer")
		return
	}

	// Se delega la decisión a la capa de dominio (estado.go)
	if err := ValidarCambioEstado(pago.Estado, pedido.Estado); err != nil {
		respuesta.Error(w, http.StatusUnprocessableEntity,
			"de "+pago.Estado+" no se puede pasar a "+pedido.Estado)
		return
	}

	pago.Estado = pedido.Estado
	if err := m.Repo.GuardarPago(&pago); err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "no se pudo guardar")
		return
	}
	respuesta.JSON(w, http.StatusOK, pago)
}

func (m *Manejador) listarPaquetes(w http.ResponseWriter, r *http.Request) {
	paquetes, err := m.Repo.ListarPaquetes()
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

	pago.ID = 0
	if pago.UsuarioID == 0 {
		pago.UsuarioID = 1
	}
	if pago.Estado == "" {
		pago.Estado = "pendiente"
	}

	if !estadosPagoValidos[pago.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "Estado de pago no permitido")
		return
	}
	if pago.PaqueteGemasID == 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "El paquete_gemas_id es obligatorio")
		return
	}
	if pago.Monto <= 0 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "El monto debe ser mayor a cero")
		return
	}

	if err := m.Repo.CrearPago(&pago); err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "No se pudo registrar el pago")
		return
	}
	respuesta.Exito(w, http.StatusCreated, pago)
}

func (m *Manejador) listarPagos(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	pagos, err := m.Repo.ListarPagos(estado)
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
	pago, err := m.Repo.BuscarPago(id)
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
	existente, err := m.Repo.BuscarPago(id)
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

	if err := m.Repo.GuardarPago(&existente); err != nil {
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
	filas, err := m.Repo.BorrarPago(id)
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
	var mision Mision
	if err := json.NewDecoder(r.Body).Decode(&mision); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}
	mision.ID = 0
	if !estadosValidos[mision.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "Estado no válido")
		return
	}
	if mision.Titulo == "" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "El título de la misión no puede estar vacío")
		return
	}
	if mision.Dificultad != "facil" && mision.Dificultad != "media" && mision.Dificultad != "dificil" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "La dificultad debe ser facil, media o dificil")
		return
	}
	if err := m.Repo.CrearMision(&mision); err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "No se pudo guardar la misión")
		return
	}
	respuesta.Exito(w, http.StatusCreated, mision)
}

func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	lista, err := m.Repo.ListarMisiones(estado)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	respuesta.Exito(w, http.StatusOK, lista)
}

func (m *Manejador) verUno(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	mision, err := m.Repo.BuscarMision(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	respuesta.Exito(w, http.StatusOK, mision)
}

func (m *Manejador) actualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	existente, err := m.Repo.BuscarMision(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "Misión no encontrada")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	var datosNuevos Mision
	if err := json.NewDecoder(r.Body).Decode(&datosNuevos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "El cuerpo no es un JSON válido")
		return
	}
	if !estadosValidos[datosNuevos.Estado] || datosNuevos.Titulo == "" ||
		(datosNuevos.Dificultad != "facil" && datosNuevos.Dificultad != "media" && datosNuevos.Dificultad != "dificil") {
		respuesta.Error(w, http.StatusUnprocessableEntity, "Datos de misión inválidos")
		return
	}
	existente.Titulo = datosNuevos.Titulo
	existente.Dificultad = datosNuevos.Dificultad
	if err := m.Repo.GuardarMision(&existente); err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "No se pudo actualizar la misión")
		return
	}
	respuesta.Exito(w, http.StatusOK, existente)
}

func (m *Manejador) borrar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	filas, err := m.Repo.BorrarMision(id)
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
