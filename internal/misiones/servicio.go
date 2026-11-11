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

func (s *Servicio) ListarPagosPorUsuario(usuarioID uint) ([]Pago, error) {
	todos, err := s.repo.ListarPagos("")
	if err != nil {
		return nil, err
	}
	var mios []Pago
	for _, p := range todos {
		if p.UsuarioID == usuarioID {
			mios = append(mios, p)
		}
	}
	return mios, nil
}

func (s *Servicio) CrearPago(usuarioID uint, peticion PagoPeticion) (Pago, error) {
	if peticion.PaqueteGemasID == 0 {
		return Pago{}, fmt.Errorf("el paquete_gemas_id es obligatorio")
	}
	if peticion.Monto <= 0 {
		return Pago{}, fmt.Errorf("el monto debe ser mayor a cero")
	}
	if usuarioID == 0 {
		usuarioID = 1
	}

	pago := peticion.AEntidad()
	pago.ID = 0
	pago.UsuarioID = usuarioID
	pago.Estado = "pendiente" // Paso 2.5: el servidor fija siempre el estado inicial

	err := s.repo.CrearPago(&pago)
	return pago, err
}

func (s *Servicio) BuscarPago(id uint) (Pago, error) {
	return s.repo.BuscarPago(id)
}

func (s *Servicio) ActualizarPago(id uint, peticion PagoPeticion) (Pago, error) {
	existente, err := s.repo.BuscarPago(id)
	if err != nil {
		return Pago{}, err
	}
	if peticion.PaqueteGemasID == 0 || peticion.Monto <= 0 {
		return Pago{}, fmt.Errorf("paquete_gemas_id y monto son obligatorios")
	}
	// Paso 2.5: se mantiene el estado y usuario_id de la fila guardada
	existente.PaqueteGemasID = peticion.PaqueteGemasID
	existente.Monto = peticion.Monto
	existente.ComprobanteURL = peticion.ComprobanteURL

	err = s.repo.GuardarPago(&existente)
	return existente, err
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

func (s *Servicio) CrearMision(peticion MisionPeticion) (Mision, error) {
	if peticion.Titulo == "" {
		return Mision{}, fmt.Errorf("el titulo de la mision no puede estar vacio")
	}
	if peticion.Dificultad != "facil" && peticion.Dificultad != "media" && peticion.Dificultad != "dificil" {
		return Mision{}, fmt.Errorf("la dificultad debe ser facil, media o dificil")
	}

	mision := peticion.AEntidad()
	mision.ID = 0
	mision.UsuarioID = 1
	mision.Estado = "pendiente"

	err := s.repo.CrearMision(&mision)
	return mision, err
}

func (s *Servicio) ActualizarMision(id uint, peticion MisionPeticion) (Mision, error) {
	existente, err := s.repo.BuscarMision(id)
	if err != nil {
		return Mision{}, err
	}
	if peticion.Titulo == "" || (peticion.Dificultad != "facil" && peticion.Dificultad != "media" && peticion.Dificultad != "dificil") {
		return Mision{}, fmt.Errorf("datos de mision invalidos")
	}
	existente.Titulo = peticion.Titulo
	existente.Categoria = peticion.Categoria
	existente.Dificultad = peticion.Dificultad
	err = s.repo.GuardarMision(&existente)
	return existente, err
}

func (s *Servicio) BorrarMision(id uint) (int64, error) {
	return s.repo.BorrarMision(id)
}

// ConfirmarPagoYEmitirFactura cambia el estado del pago a completado y emite su factura de forma atómica (Fase 6).
func (s *Servicio) ConfirmarPagoYEmitirFactura(id uint, identificacion, razonSocial string) (Pago, error) {
	pago, err := s.repo.BuscarPago(id)
	if err != nil {
		return Pago{}, err
	}
	if err := ValidarCambioEstado(pago.Estado, "completado"); err != nil {
		return Pago{}, err
	}
	pago.Estado = "completado"
	factura := Factura{
		PagoID:         id,
		Identificacion: identificacion,
		RazonSocial:    razonSocial,
	}
	if err := s.repo.GuardarPagoYFactura(&pago, &factura); err != nil {
		return Pago{}, err
	}
	return pago, nil
}
