package transporte

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/sesion"
)

type Avisos struct {
	mu      sync.Mutex
	canales map[string]map[chan misiones.Aviso]struct{}
	porRol  map[string]map[chan misiones.Aviso]struct{}
	tokens  *sesion.Tokens
	origen  string
}

func NuevosAvisos(tokens *sesion.Tokens, origen string) *Avisos {
	return &Avisos{
		canales: make(map[string]map[chan misiones.Aviso]struct{}),
		porRol:  make(map[string]map[chan misiones.Aviso]struct{}),
		tokens:  tokens,
		origen:  origen,
	}
}

func (a *Avisos) suscribir(quien, rol string) (chan misiones.Aviso, func()) {
	c := make(chan misiones.Aviso, 8)
	a.mu.Lock()
	if a.canales[quien] == nil {
		a.canales[quien] = make(map[chan misiones.Aviso]struct{})
	}
	a.canales[quien][c] = struct{}{}

	if a.porRol[rol] == nil {
		a.porRol[rol] = make(map[chan misiones.Aviso]struct{})
	}
	a.porRol[rol][c] = struct{}{}
	a.mu.Unlock()

	cancelar := func() {
		a.mu.Lock()
		if set, ok := a.canales[quien]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(a.canales, quien)
			}
		}
		if set, ok := a.porRol[rol]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(a.porRol, rol)
			}
		}
		a.mu.Unlock()
		close(c)
	}
	return c, cancelar
}

func (a *Avisos) Enviar(quien string, av misiones.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for c := range a.canales[quien] {
		select {
		case c <- av:
		default:
		}
	}
}

func (a *Avisos) EnviarAlNegocio(av misiones.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for c := range a.porRol[sesion.RolAdmin] {
		select {
		case c <- av:
		default:
		}
	}
}

func (a *Avisos) Conectar(w http.ResponseWriter, r *http.Request) {
	if a.tokens == nil {
		respuesta.Error(w, http.StatusUnauthorized, "sin sesión: falta el token o no es válido")
		return
	}
	s, err := a.tokens.Leer(r.URL.Query().Get("token"))
	if err != nil {
		respuesta.Error(w, http.StatusUnauthorized, "sin sesión: falta el token o no es válido")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{a.origen},
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	quien := strconv.Itoa(int(s.UsuarioID))
	c, cancelar := a.suscribir(quien, s.Rol)
	defer cancelar()

	ctx := conn.CloseRead(context.Background())

	for {
		select {
		case <-ctx.Done():
			return
		case av, ok := <-c:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, conn, av); err != nil {
				return
			}
		}
	}
}
