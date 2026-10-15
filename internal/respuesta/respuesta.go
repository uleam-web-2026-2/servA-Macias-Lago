package respuesta

import (
	"encoding/json"
	"net/http"
)

type detalleError struct {
	Codigo  string `json:"codigo,omitempty"`
	Mensaje string `json:"mensaje"`
}

type envoltura struct {
	OK    bool          `json:"ok"`
	Datos any           `json:"datos,omitempty"`
	Error *detalleError `json:"error,omitempty"`
}

func escribir(w http.ResponseWriter, estado int, cuerpo envoltura) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(cuerpo)
}

func Exito(w http.ResponseWriter, estado int, datos any) {
	escribir(w, estado, envoltura{OK: true, Datos: datos})
}

func JSON(w http.ResponseWriter, estado int, datos any) {
	Exito(w, estado, datos)
}

func Error(w http.ResponseWriter, estado int, textos ...string) {
	det := &detalleError{}
	if len(textos) == 1 {
		det.Mensaje = textos[0]
	} else if len(textos) >= 2 {
		det.Codigo = textos[0]
		det.Mensaje = textos[1]
	}
	escribir(w, estado, envoltura{OK: false, Error: det})
}
