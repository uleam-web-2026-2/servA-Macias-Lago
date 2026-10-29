package misiones

import (
	"errors"
	"slices"
)

// ErrTransicion indica que el cambio de estado viola las reglas del negocio.
var ErrTransicion = errors.New("transición de estado no permitida")

// ErrEstadoInvalido indica que el estado solicitado no existe en el catálogo.
var ErrEstadoInvalido = errors.New("estado no permitido")

// siguientes define la máquina de estados de Pago (Sección 5 de la Ficha).
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

// ValidarCambioEstado aplica la regla de negocio sin saber nada de HTTP ni de base de datos.
func ValidarCambioEstado(actual, nuevo string) error {
	if _, existe := siguientes[nuevo]; !existe {
		return ErrEstadoInvalido
	}
	if !puedePasar(actual, nuevo) {
		return ErrTransicion
	}
	return nil
}
