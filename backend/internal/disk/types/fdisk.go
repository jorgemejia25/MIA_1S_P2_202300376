package types

type FDisk struct {
	Path  string
	Size  int
	Unit  string
	Fit   string
	Name  string
	Type  string
	Del   string
	Start int // Inicio de la partición en el disco
	Add   int // Espacio a añadir a la partición
}
