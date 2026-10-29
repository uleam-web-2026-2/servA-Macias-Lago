package misiones

import "gorm.io/gorm"

// Repositorio define las operaciones de persistencia de la entidad principal.
type Repositorio struct {
	DB *gorm.DB
}

func NuevoRepositorio(db *gorm.DB) *Repositorio {
	return &Repositorio{DB: db}
}

func (r *Repositorio) ListarPaquetes() ([]PaqueteGemas, error) {
	var paquetes []PaqueteGemas
	err := r.DB.Find(&paquetes).Error
	return paquetes, err
}

func (r *Repositorio) ListarPagos(estado string) ([]Pago, error) {
	var pagos []Pago
	q := r.DB
	if estado != "" {
		q = q.Where("estado = ?", estado)
	}
	err := q.Order("id asc").Find(&pagos).Error
	return pagos, err
}

func (r *Repositorio) CrearPago(p *Pago) error {
	return r.DB.Create(p).Error
}

func (r *Repositorio) BuscarPago(id uint) (Pago, error) {
	var p Pago
	err := r.DB.First(&p, id).Error
	return p, err
}

func (r *Repositorio) GuardarPago(p *Pago) error {
	return r.DB.Save(p).Error
}

func (r *Repositorio) BorrarPago(id uint) (int64, error) {
	res := r.DB.Delete(&Pago{}, id)
	return res.RowsAffected, res.Error
}

// Operaciones de Misiones
func (r *Repositorio) CrearMision(m *Mision) error {
	return r.DB.Create(m).Error
}

func (r *Repositorio) ListarMisiones(estado string) ([]Mision, error) {
	var lista []Mision
	q := r.DB
	if estado != "" {
		q = q.Where("estado = ?", estado)
	}
	err := q.Find(&lista).Error
	return lista, err
}

func (r *Repositorio) BuscarMision(id uint) (Mision, error) {
	var m Mision
	err := r.DB.First(&m, id).Error
	return m, err
}

func (r *Repositorio) GuardarMision(m *Mision) error {
	return r.DB.Save(m).Error
}

func (r *Repositorio) BorrarMision(id uint) (int64, error) {
	res := r.DB.Delete(&Mision{}, id)
	return res.RowsAffected, res.Error
}
