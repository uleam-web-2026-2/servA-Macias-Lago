package misiones

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
)

// siguientes es la máquina de estados: desde cada estado, a cuáles se
// puede pasar. Van TODOS los estados como clave; los estados finales
// llevan la lista vacía. (Sección 5 de la Ficha del Negocio)
var siguientes = map[string][]string{
	"pendiente":          {"pago_por_verificar"},
	"pago_por_verificar": {"completado", "rechazado"},
	"completado":         {},
	"rechazado":          {},
}

// puedePasar dice si la máquina de estados permite ir de un estado a otro.
func puedePasar(de, a string) bool {
	return slices.Contains(siguientes[de], a)
}

// cambiarEstado: PATCH /pagos/{id} con el cuerpo {"estado": "..."}.
// 400 id o JSON · 422 estado que no existe · 404 no existe
// · 422 transición prohibida · 200 cambiado.
func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest,
			"el id debe ser un número entero positivo")
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
		respuesta.Error(w, http.StatusUnprocessableEntity,
			"estado no permitido: "+pedido.Estado)
		return
	}
	// ---- desde aquí hace falta la base de datos ----
	var t Pago
	err = m.DB.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respuesta.Error(w, http.StatusNotFound, "no existe el pago")
		return
	}
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "no se pudo leer")
		return
	}
	// La regla del negocio: se compara el estado que TIENE con el que PIDEN.
	if !puedePasar(t.Estado, pedido.Estado) {
		respuesta.Error(w, http.StatusUnprocessableEntity,
			"de "+t.Estado+" no se puede pasar a "+pedido.Estado)
		return
	}
	t.Estado = pedido.Estado
	if err := m.DB.Save(&t).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "no se pudo guardar")
		return
	}
	respuesta.JSON(w, http.StatusOK, t)
}
