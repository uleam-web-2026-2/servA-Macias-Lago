package transporte

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/sesion"
)

const origenPrueba = "http://localhost:5173"

func nuevoServidorPrueba(t *testing.T) (*httptest.Server, *sesion.Tokens, *Avisos) {
	t.Helper()
	tokens := sesion.NuevosTokens("clave-de-prueba-taller13")
	avisos := NuevosAvisos(tokens, origenPrueba)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", avisos.Conectar)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, tokens, avisos
}

func wsURL(srv *httptest.Server, token string) string {
	base := strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws"
	if token == "" {
		return base
	}
	return base + "?token=" + token
}

func TestWebSocketSinTokenRechaza401(t *testing.T) {
	srv, _, _ := nuevoServidorPrueba(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL(srv, ""), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{origenPrueba}},
	})
	if err == nil {
		t.Fatal("se esperaba rechazo sin token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		got := 0
		if resp != nil {
			got = resp.StatusCode
		}
		t.Fatalf("se esperaba 401, se obtuvo %d", got)
	}
}

func TestWebSocketOtroOrigenRechaza403(t *testing.T) {
	srv, tokens, _ := nuevoServidorPrueba(t)
	tok, err := tokens.Emitir(sesion.Usuario{ID: 1, Nombre: "ana", Rol: sesion.RolJugador})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL(srv, tok), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://malo.example"}},
	})
	if err == nil {
		t.Fatal("se esperaba rechazo desde otro origen")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		got := 0
		if resp != nil {
			got = resp.StatusCode
		}
		t.Fatalf("se esperaba 403, se obtuvo %d", got)
	}
}

func TestWebSocketAvisoLlegaSoloAlDueno(t *testing.T) {
	srv, tokens, avisos := nuevoServidorPrueba(t)

	tokAna, err := tokens.Emitir(sesion.Usuario{ID: 1, Nombre: "ana", Rol: sesion.RolJugador})
	if err != nil {
		t.Fatal(err)
	}
	tokEva, err := tokens.Emitir(sesion.Usuario{ID: 2, Nombre: "eva", Rol: sesion.RolJugador})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	connAna, _, err := websocket.Dial(ctx, wsURL(srv, tokAna), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{origenPrueba}},
	})
	if err != nil {
		t.Fatalf("no conectó ana: %v", err)
	}
	defer connAna.CloseNow()

	connEva, _, err := websocket.Dial(ctx, wsURL(srv, tokEva), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{origenPrueba}},
	})
	if err != nil {
		t.Fatalf("no conectó eva: %v", err)
	}
	defer connEva.CloseNow()

	time.Sleep(100 * time.Millisecond)

	avisos.Enviar("1", misiones.Aviso{
		Tipo:   "estado_cambiado",
		ID:     10,
		Estado: "pago_por_verificar",
	})

	recibidoAna := make(chan misiones.Aviso, 1)
	go func() {
		var av misiones.Aviso
		if err := wsjson.Read(ctx, connAna, &av); err == nil {
			recibidoAna <- av
		}
	}()

	recibidoEva := make(chan misiones.Aviso, 1)
	evaCtx, cancelEva := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancelEva()
	go func() {
		var av misiones.Aviso
		if err := wsjson.Read(evaCtx, connEva, &av); err == nil {
			recibidoEva <- av
		}
	}()

	select {
	case av := <-recibidoAna:
		if av.ID != 10 || av.Tipo != "estado_cambiado" {
			t.Fatalf("aviso inesperado para ana: %+v", av)
		}
	case <-ctx.Done():
		t.Fatal("ana no recibió su aviso a tiempo")
	}

	select {
	case av := <-recibidoEva:
		t.Fatalf("eva no debía recibir el aviso de ana, pero recibió: %+v", av)
	case <-evaCtx.Done():
	}
}