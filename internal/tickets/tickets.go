package tickets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/respuesta"
)

type Ticket struct {
	ID        int    `json:"id"`
	Titulo    string `json:"titulo"`
	Prioridad string `json:"prioridad"`
	Estado    string `json:"estado"`
}

type Almacen struct {
	mu     sync.Mutex
	ultimo int
	items  map[int]Ticket
}

func NuevoAlmacen() *Almacen {
	return &Almacen{items: map[int]Ticket{}}
}

func (a *Almacen) crear(titulo, prioridad string) Ticket {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ultimo++
	t := Ticket{ID: a.ultimo, Titulo: titulo, Prioridad: prioridad, Estado: "abierto"}
	a.items[t.ID] = t
	return t
}

func (a *Almacen) obtener(id int) (Ticket, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t, ok := a.items[id]
	return t, ok
}

func (a *Almacen) listar() []Ticket {
	a.mu.Lock()
	defer a.mu.Unlock()
	lista := make([]Ticket, 0, len(a.items))
	for i := 1; i <= a.ultimo; i++ {
		if t, ok := a.items[i]; ok {
			lista = append(lista, t)
		}
	}
	return lista
}

func (a *Almacen) Listar(w http.ResponseWriter, r *http.Request) {
	respuesta.Exito(w, http.StatusOK, a.listar())
}

func (a *Almacen) Obtener(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "el id debe ser un entero positivo")
		return
	}
	t, ok := a.obtener(id)
	if !ok {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "no existe un ticket con ese id")
		return
	}
	respuesta.Exito(w, http.StatusOK, t)
}

type entradaCrear struct {
	Titulo    string `json:"titulo"`
	Prioridad string `json:"prioridad"`
}

func (a *Almacen) Crear(w http.ResponseWriter, r *http.Request) {
	var entrada entradaCrear
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "cuerpo_malformado", "el cuerpo debe ser JSON")
		return
	}
	t := a.crear(entrada.Titulo, entrada.Prioridad)
	respuesta.Exito(w, http.StatusCreated, t)
}

func Explotar(w http.ResponseWriter, r *http.Request) {
	var t *Ticket
	_ = t.Titulo
}