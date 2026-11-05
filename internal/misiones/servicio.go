package misiones

import "fmt"

type Servicio struct {
	repo Repositorio
}

func NuevoServicio(repo Repositorio) *Servicio {
	return &Servicio{repo: repo}
}

func (s *Servicio) ListarPaquetes() ([]PaqueteGemas, error) {
	return s.repo.ListarPaquetes()
}

func (s *Servicio) ListarPagos(estado string) ([]Pago, error) {
	return s.repo.ListarPagos(estado)
}

func (s *Servicio) CrearPago(pago Pago) (Pago, error) {
	pago.ID = 0
	if pago.UsuarioID == 0 {
		pago.UsuarioID = 1
	}
	if pago.Estado == "" {
		pago.Estado = "pendiente"
	}
	if !estadosPagoValidos[pago.Estado] {
		return pago, fmt.Errorf("estado de pago no permitido")
	}
	if pago.PaqueteGemasID == 0 {
		return pago, fmt.Errorf("el paquete_gemas_id es obligatorio")
	}
	if pago.Monto <= 0 {
		return pago, fmt.Errorf("el monto debe ser mayor a cero")
	}
	err := s.repo.CrearPago(&pago)
	return pago, err
}

func (s *Servicio) BuscarPago(id uint) (Pago, error) {
	return s.repo.BuscarPago(id)
}

func (s *Servicio) GuardarPago(pago *Pago) error {
	return s.repo.GuardarPago(pago)
}

func (s *Servicio) BorrarPago(id uint) (int64, error) {
	return s.repo.BorrarPago(id)
}

func (s *Servicio) CambiarEstadoPago(id uint, nuevoEstado string) (Pago, error) {
	if _, existe := siguientes[nuevoEstado]; !existe {
		return Pago{}, fmt.Errorf("estado no permitido: %s", nuevoEstado)
	}
	pago, err := s.repo.BuscarPago(id)
	if err != nil {
		return Pago{}, err
	}
	if err := ValidarCambioEstado(pago.Estado, nuevoEstado); err != nil {
		return Pago{}, err
	}
	pago.Estado = nuevoEstado
	err = s.repo.GuardarPago(&pago)
	return pago, err
}

// Metodos para Misiones
func (s *Servicio) ListarMisiones(estado string) ([]Mision, error) {
	return s.repo.ListarMisiones(estado)
}

func (s *Servicio) BuscarMision(id uint) (Mision, error) {
	return s.repo.BuscarMision(id)
}

func (s *Servicio) CrearMision(mision Mision) (Mision, error) {
	mision.ID = 0
	if !estadosValidos[mision.Estado] {
		return mision, fmt.Errorf("estado no valido")
	}
	if mision.Titulo == "" {
		return mision, fmt.Errorf("el titulo de la mision no puede estar vacio")
	}
	if mision.Dificultad != "facil" && mision.Dificultad != "media" && mision.Dificultad != "dificil" {
		return mision, fmt.Errorf("la dificultad debe ser facil, media o dificil")
	}
	err := s.repo.CrearMision(&mision)
	return mision, err
}

func (s *Servicio) ActualizarMision(id uint, cambios Mision) (Mision, error) {
	existente, err := s.repo.BuscarMision(id)
	if err != nil {
		return Mision{}, err
	}
	if !estadosValidos[cambios.Estado] || cambios.Titulo == "" ||
		(cambios.Dificultad != "facil" && cambios.Dificultad != "media" && cambios.Dificultad != "dificil") {
		return Mision{}, fmt.Errorf("datos de mision invalidos")
	}
	existente.Titulo = cambios.Titulo
	existente.Dificultad = cambios.Dificultad
	err = s.repo.GuardarMision(&existente)
	return existente, err
}

func (s *Servicio) BorrarMision(id uint) (int64, error) {
	return s.repo.BorrarMision(id)
}