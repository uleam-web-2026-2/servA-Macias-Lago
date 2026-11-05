package misiones

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNoExiste = errors.New("registro no encontrado")

// Repositorio es la interfaz que abstrae la base de datos para la capa de servicio.
type Repositorio interface {
	ListarPaquetes() ([]PaqueteGemas, error)
	ListarPagos(estado string) ([]Pago, error)
	CrearPago(p *Pago) error
	BuscarPago(id uint) (Pago, error)
	GuardarPago(p *Pago) error
	BorrarPago(id uint) (int64, error)

	CrearMision(m *Mision) error
	ListarMisiones(estado string) ([]Mision, error)
	BuscarMision(id uint) (Mision, error)
	GuardarMision(m *Mision) error
	BorrarMision(id uint) (int64, error)
}

// AlmacenGorm implementa Repositorio usando GORM.
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
	return p, err
}

func (a *AlmacenGorm) GuardarPago(p *Pago) error {
	return a.DB.Save(p).Error
}

func (a *AlmacenGorm) BorrarPago(id uint) (int64, error) {
	res := a.DB.Delete(&Pago{}, id)
	return res.RowsAffected, res.Error
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
	return m, err
}

func (a *AlmacenGorm) GuardarMision(m *Mision) error {
	return a.DB.Save(m).Error
}

func (a *AlmacenGorm) BorrarMision(id uint) (int64, error) {
	res := a.DB.Delete(&Mision{}, id)
	return res.RowsAffected, res.Error
}