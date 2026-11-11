package misiones

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func enrutador() http.Handler {
	return enrutadorCon(rolDePrueba)
}

func TestCrear(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		codigo int
	}{
		{
			nombre: "JSON roto responde 400",
			cuerpo: `{"titulo": "Derrotar al Jefe",`,
			codigo: http.StatusBadRequest,
		},
		{
			nombre: "titulo vacio responde 422",
			cuerpo: `{"titulo": "", "categoria": "Diaria", "dificultad": "facil"}`,
			codigo: http.StatusUnprocessableEntity,
		},
		{
			nombre: "dificultad no permitida responde 422 (Regla propia)",
			cuerpo: `{"titulo": "Derrotar al Jefe", "categoria": "Diaria", "dificultad": "legendaria"}`,
			codigo: http.StatusUnprocessableEntity,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			peticion := httptest.NewRequest(http.MethodPost, "/misiones", strings.NewReader(caso.cuerpo))
			grabadora := httptest.NewRecorder()

			enrutador().ServeHTTP(grabadora, peticion)

			if grabadora.Code != caso.codigo {
				t.Fatalf("se esperaba %d y llegó %d con cuerpo %s", caso.codigo, grabadora.Code, grabadora.Body.String())
			}
		})
	}
}

func TestVerUnoConIDQueNoEsNumeroResponde400(t *testing.T) {
	peticion := httptest.NewRequest(http.MethodGet, "/misiones/abc", nil)
	grabadora := httptest.NewRecorder()

	enrutador().ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 y llegó %d", grabadora.Code)
	}
}

// repoFalsoPago simula la base en memoria para probar el rollback de la transacción (Fase 6).
type repoFalsoPago struct {
	*AlmacenGorm
	pago     Pago
	facturas int
}

func (f *repoFalsoPago) BuscarPago(id uint) (Pago, error) {
	if f.pago.ID != id {
		return Pago{}, ErrNoExiste
	}
	return f.pago, nil
}

func (f *repoFalsoPago) GuardarPagoYFactura(p *Pago, fac *Factura) error {
	if fac.Identificacion == "" {
		return errors.New("error forzado: identificacion vacia rechazada por CHECK")
	}
	f.pago = *p
	f.facturas++
	return nil
}

func TestTransaccionErrorForzadoDeshaceCambioDeEstado(t *testing.T) {
	falso := &repoFalsoPago{
		pago: Pago{ID: 1, UsuarioID: 1, PaqueteGemasID: 1, Monto: 4.99, Estado: "pago_por_verificar"},
	}
	svc := NuevoServicio(falso)

	_, err := svc.ConfirmarPagoYEmitirFactura(1, "", "Cliente Prueba")
	if err == nil {
		t.Fatal("se esperaba error al forzar identificacion vacia, pero paso sin error")
	}
	if falso.pago.Estado != "pago_por_verificar" {
		t.Fatalf("el fallo a mitad de camino no deshizo el cambio: estado quedo en %s", falso.pago.Estado)
	}
	if falso.facturas != 0 {
		t.Fatalf("se esperaban 0 facturas y quedaron %d", falso.facturas)
	}
}
