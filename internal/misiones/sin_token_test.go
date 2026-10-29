package misiones

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSinTokenLasRutasProtegidasResponden401(t *testing.T) {
	protegidas := []struct{ metodo, ruta, cuerpo string }{
		{"POST", "/misiones", `{"titulo":"Prueba","categoria":"Diaria","dificultad":"facil","estado":"pendiente"}`},
		{"PUT", "/misiones/1", `{"titulo":"Prueba","categoria":"Diaria","dificultad":"facil","estado":"pendiente"}`},
		{"DELETE", "/misiones/1", ""},
		{"GET", "/pagos", ""},
		{"GET", "/pagos/mios", ""},
		{"POST", "/pagos", `{"usuario_id":1,"paquete_gemas_id":1,"monto":4.99}`},
		{"GET", "/pagos/1", ""},
		{"PUT", "/pagos/1", `{"usuario_id":1,"paquete_gemas_id":1,"monto":4.99}`},
		{"DELETE", "/pagos/1", ""},
		{"PATCH", "/pagos/1", `{"estado":"pago_por_verificar"}`},
	}

	servidor := enrutadorCon("")

	for _, p := range protegidas {
		peticion := httptest.NewRequest(p.metodo, p.ruta, strings.NewReader(p.cuerpo))
		grabadora := httptest.NewRecorder()

		cayo := seCayo(t, func() { servidor.ServeHTTP(grabadora, peticion) })

		if cayo {
			t.Errorf("%s %s sin token: el guardia no la detuvo y llegó a la base", p.metodo, p.ruta)
			continue
		}
		if grabadora.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin token: se esperaba 401 y llegó %d", p.metodo, p.ruta, grabadora.Code)
		}
	}
}
