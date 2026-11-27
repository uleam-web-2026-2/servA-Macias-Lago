package misiones

type repoFalso struct {
	Repositorio
	pago Pago
}

func (f *repoFalso) BuscarPago(id uint) (Pago, error) {
	if f.pago.ID != id {
		return Pago{}, ErrNoExiste
	}
	return f.pago, nil
}

func (f *repoFalso) GuardarPago(p *Pago) error {
	f.pago = *p
	return nil
}