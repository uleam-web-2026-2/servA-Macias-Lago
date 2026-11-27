package misiones

import (
	"strconv"
	"testing"
)

type avisadorFalso struct {
	quien  []string
	avisos []Aviso
}

func (a *avisadorFalso) Enviar(quien string, av Aviso) {
	a.quien = append(a.quien, quien)
	a.avisos = append(a.avisos, av)
}

func TestCambiarEstadoAvisaAlDueno(t *testing.T) {
	f := &repoFalso{pago: Pago{ID: 7, UsuarioID: 1, Monto: 4.99, Estado: "pendiente"}}
	av := &avisadorFalso{}
	s := NuevoServicio(f)
	s.AvisarCon(av)

	if _, err := s.CambiarEstadoPago(7, "pago_por_verificar"); err != nil {
		t.Fatal(err)
	}
	esperadoQuien := strconv.Itoa(1)
	if len(av.quien) != 1 || av.quien[0] != esperadoQuien {
		t.Fatalf("se esperaba un aviso para %s y se avisó a %v", esperadoQuien, av.quien)
	}
	if av.avisos[0].ID != 7 || av.avisos[0].Tipo == "" {
		t.Fatalf("el aviso tiene que traer tipo e id: %+v", av.avisos[0])
	}
}

func TestTransicionProhibidaNoAvisa(t *testing.T) {
	f := &repoFalso{pago: Pago{ID: 7, UsuarioID: 1, Monto: 4.99, Estado: "pendiente"}}
	av := &avisadorFalso{}
	s := NuevoServicio(f)
	s.AvisarCon(av)

	if _, err := s.CambiarEstadoPago(7, "completado"); err == nil {
		t.Fatal("de pendiente a completado tenía que fallar")
	}
	if len(av.quien) != 0 {
		t.Fatalf("no cambió nada y se avisó a %v", av.quien)
	}
}