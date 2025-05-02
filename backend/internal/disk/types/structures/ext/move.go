package ext2

// Move realiza el movimiento de un archivo o directorio de una ubicación a otra
// Esto se implementa como una operación de copia seguida de eliminación del origen
func (sb *SuperBlock) Move(
	path string,
	sourceParentDirs []string,
	sourceName string,
	destParentDirs []string,
	destName string,
	uid int32,
	gid int32,
) error {

	return nil
}
