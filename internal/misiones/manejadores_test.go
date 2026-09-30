package misiones

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func enrutador() http.Handler {
	r := chi.NewRouter()
	(&Manejador{DB: nil}).Rutas(r)
	return r
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
			cuerpo: `{"titulo": "", "categoria": "Diaria", "dificultad": "facil", "estado": "pendiente"}`,
			codigo: http.StatusUnprocessableEntity,
		},
		{
			nombre: "estado invalido responde 422",
			cuerpo: `{"titulo": "Derrotar al Jefe", "categoria": "Diaria", "dificultad": "facil", "estado": "invalido"}`,
			codigo: http.StatusUnprocessableEntity,
		},
		{
			nombre: "dificultad no permitida responde 422 (Regla propia)",
			cuerpo: `{"titulo": "Derrotar al Jefe", "categoria": "Diaria", "dificultad": "legendaria", "estado": "pendiente"}`,
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