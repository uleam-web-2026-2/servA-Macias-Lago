package misiones

type Aviso struct {
	Tipo   string `json:"tipo"`
	ID     uint   `json:"id"`
	Estado string `json:"estado"`
}

type Avisador interface {
	Enviar(quien string, av Aviso)
	EnviarAlNegocio(av Aviso)
}