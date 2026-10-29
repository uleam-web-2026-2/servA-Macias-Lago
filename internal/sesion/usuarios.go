package sesion

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Roles del negocio DoubleLevel (Sección 6 de la Ficha)
const (
	RolJugador = "jugador"
	RolAdmin   = "admin"
)

type Usuario struct {
	ID        uint   `json:"id"`
	Nombre    string `json:"nombre" gorm:"uniqueIndex"`
	Rol       string `json:"rol"`
	ClaveHash string `json:"-"`
}

// Evita que choque con la tabla usuarios del juego en modelos.go
func (Usuario) TableName() string {
	return "cuentas"
}

func Sembrar(db *gorm.DB) error {
	ejemplos := []struct{ nombre, rol, clave string }{
		{"ana", RolJugador, "ana-2026"},
		{"luis", RolAdmin, "luis-2026"},
	}
	for _, e := range ejemplos {
		resumen, err := bcrypt.GenerateFromPassword(
			[]byte(e.clave), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		var u Usuario
		db.Where("nombre = ?", e.nombre).Limit(1).Find(&u)
		u.Nombre, u.Rol, u.ClaveHash = e.nombre, e.rol, string(resumen)
		if err := db.Save(&u).Error; err != nil {
			return err
		}
	}
	return nil
}

type Manejador struct {
	DB     *gorm.DB
	Tokens *Tokens
}

func (m *Manejador) Login(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		responder(w, http.StatusBadRequest, "el cuerpo no es JSON válido")
		return
	}
	const noCoincide = "usuario o contraseña incorrectos"

	var u Usuario
	err := m.DB.Where("nombre = ?", pedido.Usuario).Limit(1).Find(&u).Error
	if err != nil {
		responder(w, http.StatusInternalServerError, "no se pudo leer")
		return
	}
	if u.ID == 0 || bcrypt.CompareHashAndPassword(
		[]byte(u.ClaveHash), []byte(pedido.Clave)) != nil {
		responder(w, http.StatusUnauthorized, noCoincide)
		return
	}
	token, err := m.Tokens.Emitir(u)
	if err != nil {
		responder(w, http.StatusInternalServerError,
			"no se pudo emitir el token")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
