package respuesta

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	if datos != nil {
		_ = json.NewEncoder(w).Encode(datos)
	}
}

func Exito(w http.ResponseWriter, estado int, datos any) {
	JSON(w, estado, datos)
}

func Error(w http.ResponseWriter, estado int, textos ...string) {
	mensaje := "error en la petición"
	if len(textos) == 1 {
		mensaje = textos[0]
	} else if len(textos) >= 2 {
		mensaje = textos[1]
	}
	JSON(w, estado, map[string]string{"error": mensaje})
}
