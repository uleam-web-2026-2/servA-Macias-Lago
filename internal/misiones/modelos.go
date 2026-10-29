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

// Mision representa la entidad con estados del lado de los MUCHOS
type Mision struct {
	gorm.Model
	Titulo     string `json:"Titulo"`
	Categoria  string `json:"Categoria"`
	Dificultad string `json:"Dificultad"`
	Estado     string `json:"Estado"`
	UsuarioID  uint   `json:"UsuarioID"` // Clave foránea
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

// Pago es la entidad con máquina de estados de la Sección 5 de la Ficha
type Pago struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UsuarioID      uint         `gorm:"not null" json:"usuario_id"`
	PaqueteGemasID uint         `gorm:"not null" json:"paquete_gemas_id"`
	Monto          float64      `gorm:"type:decimal(10,2);not null" json:"monto"`
	Estado         string       `gorm:"type:varchar(30);not null;default:'pendiente'" json:"estado"` // pendiente | pago_por_verificar | completado | rechazado
	ComprobanteURL *string      `gorm:"type:text" json:"comprobante_url,omitempty"`
	FechaCreacion  time.Time    `gorm:"autoCreateTime" json:"fecha_creacion"`
	Usuario        Usuario      `gorm:"foreignKey:UsuarioID" json:"-"`
	PaqueteGemas   PaqueteGemas `gorm:"foreignKey:PaqueteGemasID" json:"-"`
}

type Factura struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PagoID         uint      `gorm:"uniqueIndex;not null" json:"pago_id"`
	Identificacion string    `gorm:"type:varchar(20);not null" json:"identificacion"`
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

// Estados válidos de Pago según la Sección 5 de la Ficha
var estadosPagoValidos = map[string]bool{
	"pendiente":          true,
	"pago_por_verificar": true,
	"completado":         true,
	"rechazado":          true,
}
