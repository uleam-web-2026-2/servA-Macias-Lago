package misiones

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/uleam-web-2026-2/servA-Macias-Lago/internal/misiones/sqlc"
)

var ErrNoExiste = errors.New("registro no encontrado")

type Repositorio interface {
	ListarPaquetes() ([]PaqueteGemas, error)
	ListarPagos(estado string) ([]Pago, error)
	CrearPago(p *Pago) error
	BuscarPago(id uint) (Pago, error)
	GuardarPago(p *Pago) error
	BorrarPago(id uint) (int64, error)
	GuardarPagoYFactura(p *Pago, f *Factura) error

	CrearMision(m *Mision) error
	ListarMisiones(estado string) ([]Mision, error)
	BuscarMision(id uint) (Mision, error)
	GuardarMision(m *Mision) error
	BorrarMision(id uint) (int64, error)
}

type AlmacenGorm struct {
	DB *gorm.DB
}

var _ Repositorio = (*AlmacenGorm)(nil)

func NuevoAlmacenGorm(db *gorm.DB) *AlmacenGorm {
	return &AlmacenGorm{DB: db}
}

func (a *AlmacenGorm) ListarPaquetes() ([]PaqueteGemas, error) {
	var paquetes []PaqueteGemas
	err := a.DB.Find(&paquetes).Error
	return paquetes, err
}

func (a *AlmacenGorm) ListarPagos(estado string) ([]Pago, error) {
	var pagos []Pago
	q := a.DB
	if estado != "" {
		q = q.Where("estado = ?", estado)
	}
	err := q.Order("id asc").Find(&pagos).Error
	return pagos, err
}

func (a *AlmacenGorm) CrearPago(p *Pago) error {
	return a.DB.Create(p).Error
}

func (a *AlmacenGorm) BuscarPago(id uint) (Pago, error) {
	var p Pago
	err := a.DB.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, ErrNoExiste
	}
	return p, err
}

func (a *AlmacenGorm) GuardarPago(p *Pago) error {
	return a.DB.Save(p).Error
}

func (a *AlmacenGorm) BorrarPago(id uint) (int64, error) {
	res := a.DB.Delete(&Pago{}, id)
	return res.RowsAffected, res.Error
}

// GuardarPagoYFactura ejecuta las dos escrituras dentro de una transacción GORM (Fase 6).
func (a *AlmacenGorm) GuardarPagoYFactura(p *Pago, f *Factura) error {
	return a.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(p).Error; err != nil {
			return err
		}
		f.PagoID = p.ID
		return tx.Create(f).Error
	})
}

func (a *AlmacenGorm) CrearMision(m *Mision) error {
	return a.DB.Create(m).Error
}

func (a *AlmacenGorm) ListarMisiones(estado string) ([]Mision, error) {
	var lista []Mision
	q := a.DB
	if estado != "" {
		q = q.Where("estado = ?", estado)
	}
	err := q.Find(&lista).Error
	return lista, err
}

func (a *AlmacenGorm) BuscarMision(id uint) (Mision, error) {
	var m Mision
	err := a.DB.First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, ErrNoExiste
	}
	return m, err
}

func (a *AlmacenGorm) GuardarMision(m *Mision) error {
	return a.DB.Save(m).Error
}

func (a *AlmacenGorm) BorrarMision(id uint) (int64, error) {
	res := a.DB.Delete(&Mision{}, id)
	return res.RowsAffected, res.Error
}

// PagosSqlc incrusta AlmacenGorm y migra BuscarPago, GuardarPago y GuardarPagoYFactura a sqlc (Fases 3 y 6).
type PagosSqlc struct {
	*AlmacenGorm
	Conexion *sql.DB
	q        *sqlc.Queries
}

var _ Repositorio = (*PagosSqlc)(nil)

func NuevoPagosSqlc(g *AlmacenGorm, c *sql.DB) *PagosSqlc {
	return &PagosSqlc{
		AlmacenGorm: g,
		Conexion:    c,
		q:           sqlc.New(c),
	}
}

// Mantenemos NuevoMisionesSqlc como alias por compatibilidad
func NuevoMisionesSqlc(g *AlmacenGorm, c *sql.DB) *PagosSqlc {
	return NuevoPagosSqlc(g, c)
}

func (a *PagosSqlc) BuscarPago(id uint) (Pago, error) {
	f, err := a.q.BuscarPagoSqlc(context.Background(), int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return Pago{}, ErrNoExiste
	}
	if err != nil {
		return Pago{}, err
	}
	var comp *string
	if f.ComprobanteUrl.Valid {
		s := f.ComprobanteUrl.String
		comp = &s
	}
	return Pago{
		ID:             uint(f.ID),
		UsuarioID:      uint(f.UsuarioID),
		PaqueteGemasID: uint(f.PaqueteGemasID),
		Monto:          f.Monto,
		Estado:         f.Estado,
		ComprobanteURL: comp,
		FechaCreacion:  f.FechaCreacion,
	}, nil
}

func (a *PagosSqlc) GuardarPago(p *Pago) error {
	var comp sql.NullString
	if p.ComprobanteURL != nil {
		comp = sql.NullString{String: *p.ComprobanteURL, Valid: true}
	}
	n, err := a.q.GuardarPagoSqlc(context.Background(), sqlc.GuardarPagoSqlcParams{
		ID:             int64(p.ID),
		UsuarioID:      int64(p.UsuarioID),
		PaqueteGemasID: int64(p.PaqueteGemasID),
		Monto:          p.Monto,
		Estado:         p.Estado,
		ComprobanteUrl: comp,
	})
	if err == nil && n == 0 {
		return ErrNoExiste
	}
	return err
}

// GuardarPagoYFactura ejecuta ambas consultas de sqlc dentro de una transacción SQL con WithTx (Fase 6).
func (a *PagosSqlc) GuardarPagoYFactura(p *Pago, f *Factura) error {
	tx, err := a.Conexion.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := a.q.WithTx(tx)
	var comp sql.NullString
	if p.ComprobanteURL != nil {
		comp = sql.NullString{String: *p.ComprobanteURL, Valid: true}
	}
	n, err := qtx.GuardarPagoSqlc(context.Background(), sqlc.GuardarPagoSqlcParams{
		ID:             int64(p.ID),
		UsuarioID:      int64(p.UsuarioID),
		PaqueteGemasID: int64(p.PaqueteGemasID),
		Monto:          p.Monto,
		Estado:         p.Estado,
		ComprobanteUrl: comp,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoExiste
	}

	err = qtx.CrearFacturaSqlc(context.Background(), sqlc.CrearFacturaSqlcParams{
		PagoID:         int64(p.ID),
		Identificacion: f.Identificacion,
		RazonSocial:    f.RazonSocial,
	})
	if err != nil {
		return err
	}
	f.PagoID = p.ID

	return tx.Commit()
}

func (a *PagosSqlc) BuscarMision(id uint) (Mision, error) {
	f, err := a.q.BuscarMisionSqlc(context.Background(), int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return Mision{}, ErrNoExiste
	}
	if err != nil {
		return Mision{}, err
	}
	var m Mision
	m.ID = uint(f.ID)
	m.Titulo = f.Titulo
	m.Categoria = f.Categoria
	m.Dificultad = f.Dificultad
	m.Estado = f.Estado
	m.UsuarioID = uint(f.UsuarioID)
	if f.CreatedAt.Valid {
		m.CreatedAt = f.CreatedAt.Time
	}
	return m, nil
}

func (a *PagosSqlc) GuardarMision(m *Mision) error {
	n, err := a.q.GuardarMisionSqlc(context.Background(), sqlc.GuardarMisionSqlcParams{
		ID:         int64(m.ID),
		Titulo:     m.Titulo,
		Categoria:  m.Categoria,
		Dificultad: m.Dificultad,
		Estado:     m.Estado,
	})
	if err == nil && n == 0 {
		return ErrNoExiste
	}
	return err
}
