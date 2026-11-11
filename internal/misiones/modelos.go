package misiones

import (
	"time"

	"gorm.io/gorm"
)

// Usuario representa la entidad del lado del UNO
type Usuario struct {
	gorm.Model
	Nombre   string   `json:"Nombre"`
	Correo   string   `json:"correo"`
	Rol      string   `gorm:"default:'jugador'" json:"rol"` // jugador | admin
	Clase    string   `json:"Clase"`
	Nivel    int      `json:"Nivel"`
	Misiones []Mision `json:"Misiones"` // Relación 1 a N
}

// Mision representa la entidad secundaria sin etiquetas json
type Mision struct {
	gorm.Model
	Titulo     string
	Categoria  string
	Dificultad string
	Estado     string
	UsuarioID  uint
}

var estadosValidos = map[string]bool{
	"pendiente":  true,
	"completada": true,
	"fallida":    true,
}

// --- ENTIDADES DE MONETIZACIÓN (FICHA DEL NEGOCIO - HITO 1) ---

type PaqueteGemas struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	Nombre        string  `gorm:"type:varchar(100);not null" json:"nombre"`
	Precio        float64 `gorm:"type:decimal(10,2);not null" json:"precio"`
	CantidadGemas int     `gorm:"not null" json:"cantidad_gemas"`
}

// Pago es la entidad principal con máquina de estados (SIN etiquetas json ni json:"-")
type Pago struct {
	ID             uint         `gorm:"primaryKey"`
	UsuarioID      uint         `gorm:"not null"`
	PaqueteGemasID uint         `gorm:"not null"`
	Monto          float64      `gorm:"type:decimal(10,2);not null;check:monto > 0"`
	Estado         string       `gorm:"type:varchar(30);not null;default:'pendiente'"`
	ComprobanteURL *string      `gorm:"type:text"`
	FechaCreacion  time.Time    `gorm:"autoCreateTime"`
	Usuario        Usuario      `gorm:"foreignKey:UsuarioID"`
	PaqueteGemas   PaqueteGemas `gorm:"foreignKey:PaqueteGemasID"`
}

type Factura struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PagoID         uint      `gorm:"uniqueIndex;not null" json:"pago_id"`
	Identificacion string    `gorm:"type:varchar(20);not null;check:identificacion <> ''" json:"identificacion"`
	RazonSocial    string    `gorm:"type:varchar(150);not null" json:"razon_social"`
	FechaEmision   time.Time `gorm:"autoCreateTime" json:"fecha_emision"`
	Pago           Pago      `gorm:"foreignKey:PagoID" json:"-"`
}

type ConsumoGemas struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UsuarioID     uint      `gorm:"not null" json:"usuario_id"`
	GemasGastadas int       `gorm:"not null" json:"gemas_gastadas"`
	ItemObtenido  string    `gorm:"type:varchar(100);not null" json:"item_obtenido"`
	Fecha         time.Time `gorm:"autoCreateTime" json:"fecha"`
	Usuario       Usuario   `gorm:"foreignKey:UsuarioID" json:"-"`
}

var estadosPagoValidos = map[string]bool{
	"pendiente":          true,
	"pago_por_verificar": true,
	"completado":         true,
	"rechazado":          true,
}
