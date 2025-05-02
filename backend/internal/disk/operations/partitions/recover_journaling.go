package partition_operations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
)

// RecoverFromJournaling recupera archivos y carpetas ejecutando las operaciones registradas en el journaling
func RecoverFromJournaling(id string) (string, error) {
	var output strings.Builder

	// Obtener la partición montada
	partition, path, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición: %v", err)
	}

	// Leer el superbloque
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(path, partition.Partition.Part_start)
	if err != nil {
		return "", fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Verificar si es ext3 (tiene journaling)
	if superBlock.SFilesystemType != 3 {
		return "", fmt.Errorf("la partición no tiene journaling (no es ext3)")
	}

	// El inicio del journaling es justo después del SuperBlock
	journalStart := partition.Partition.Part_start + ext2.SuperBlockSize

	// Obtener todas las entradas del journal
	journals, err := ext2.GetJournaling(path, int64(journalStart), superBlock.SFreeInodesCount)
	if err != nil {
		return "", fmt.Errorf("error al obtener el journaling: %v", err)
	}

	if len(journals) == 0 {
		return "", fmt.Errorf("no hay operaciones registradas en el journaling para recuperar")
	}

	// Verificar si hay un usuario conectado para ejecutar las operaciones
	instance := auth.GetInstance()
	if instance.User == nil {
		return "", fmt.Errorf("error al recuperar archivos: no hay un usuario loggeado")
	}

	output.WriteString(fmt.Sprintf("Comenzando recuperación de %d operaciones desde el journaling...\n", len(journals)))

	// Primero, asegurar que exista la carpeta raíz (siempre debe existir)
	output.WriteString("Fase 0: Verificando estructura base del sistema...\n")

	// Verificar y crear la carpeta raíz si es necesario
	// Esto es crítico ya que todos los demás archivos y directorios dependen de la raíz
	output.WriteString("Verificando carpeta raíz '/'...\n")
	err = ensureRootDirectory(path, partition.Partition.Part_start)
	if err != nil {
		output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al asegurar la carpeta raíz: %v\n", err))
	} else {
		output.WriteString("  ✓ Carpeta raíz '/' verificada\n")
	}

	// Verificar y crear el archivo users.txt si es necesario
	output.WriteString("Verificando archivo essential 'users.txt'...\n")
	err = ensureUsersFile(path, partition.Partition.Part_start)
	if err != nil {
		output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al asegurar el archivo users.txt: %v\n", err))
	} else {
		output.WriteString("  ✓ Archivo 'users.txt' verificado\n")
	}

	// Primero, procesar todas las operaciones de creación de directorios
	// para asegurarse de que la estructura de directorios esté lista
	output.WriteString("\nFase 1: Creando estructura de directorios...\n")
	for i, journal := range journals {
		operation := strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00")
		filePath := strings.TrimRight(string(journal.J_content.I_path[:]), "\x00")

		if operation == "mkdir" {
			output.WriteString(fmt.Sprintf("[Paso 1/%d] Creando directorio '%s'...\n",
				i+1, filePath))

			// Crear el directorio con la opción recursiva
			err := CreateDirectory(filePath, true)
			if err != nil {
				output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al crear directorio '%s': %v\n", filePath, err))
			} else {
				output.WriteString(fmt.Sprintf("  ✓ Directorio base creado: %s\n", filePath))
			}
		}
	}

	// Luego, garantizar que exista la estructura de directorios necesaria
	// para los archivos que se intentarán crear
	output.WriteString("\nFase 2: Asegurando rutas para archivos...\n")
	directoriesNeeded := make(map[string]bool)

	for _, journal := range journals {
		operation := strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00")
		filePath := strings.TrimRight(string(journal.J_content.I_path[:]), "\x00")

		if operation == "mkfile" {
			// Obtener la ruta del directorio padre
			parentPath := filepath.Dir(filePath)
			if parentPath != "/" {
				directoriesNeeded[parentPath] = true
			}
		}
	}

	// Crear todos los directorios padre necesarios
	for dirPath := range directoriesNeeded {
		output.WriteString(fmt.Sprintf("Asegurando directorio: %s\n", dirPath))
		err := CreateDirectory(dirPath, true)
		if err != nil {
			output.WriteString(fmt.Sprintf("  ADVERTENCIA: No se pudo crear el directorio '%s': %v\n", dirPath, err))
		}
	}

	// Finalmente, procesar todas las operaciones de archivo
	output.WriteString("\nFase 3: Recuperando archivos y aplicando otras operaciones...\n")
	for i, journal := range journals {
		operation := strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00")
		filePath := strings.TrimRight(string(journal.J_content.I_path[:]), "\x00")
		content := strings.TrimRight(string(journal.J_content.I_content[:]), "\x00")

		// Fecha como referencia
		date := time.Unix(int64(journal.J_content.I_date), 0)

		output.WriteString(fmt.Sprintf("[%d/%d] Procesando operación '%s' en '%s' (registrada: %s)...\n",
			i+1, len(journals), operation, filePath, date.Format("02/01/2006 15:04:05")))

		// Ejecutar la operación según su tipo
		switch operation {
		case "mkdir":
			// Ya procesado en la fase 1, solo registrar
			output.WriteString(fmt.Sprintf("  ✓ Directorio ya procesado: %s\n", filePath))

		case "mkfile":
			// Determinar el tamaño del archivo (por defecto 0)
			size := 0

			// Intentar crear el archivo
			err := CreateFile(filePath, size, "", true)
			if err != nil {
				output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al recrear archivo '%s': %v\n", filePath, err))
			} else {
				// Si hay contenido en la entrada del journal, intentamos editar el archivo
				if content != "" {
					// Intentar escribir el contenido
					err = EditFile(filePath, content)
					if err != nil {
						output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al restaurar contenido de '%s': %v\n", filePath, err))
					}
				}
				output.WriteString(fmt.Sprintf("  ✓ Archivo recuperado: %s\n", filePath))
			}

		case "remove":
			// No hacemos nada en caso de eliminación, ya que queremos recuperar archivos
			output.WriteString(fmt.Sprintf("  ✓ Ignorando operación de eliminación para '%s'\n", filePath))

		case "edit":
			// Intentar editar el contenido si el archivo existe
			err = EditFile(filePath, content)
			if err != nil {
				output.WriteString(fmt.Sprintf("  ADVERTENCIA: Error al editar '%s': %v\n", filePath, err))
			} else {
				output.WriteString(fmt.Sprintf("  ✓ Contenido editado para '%s'\n", filePath))
			}

		case "rename", "chmod", "chown", "copy", "move":
			// Operaciones avanzadas
			output.WriteString(fmt.Sprintf("  ⚠ Operación '%s' no implementada en la recuperación\n", operation))

		default:
			output.WriteString(fmt.Sprintf("  ⚠ Operación '%s' desconocida\n", operation))
		}
	}

	output.WriteString(fmt.Sprintf("\nRecuperación completada. Se procesaron %d operaciones del journaling.\n", len(journals)))
	return output.String(), nil
}

// ensureRootDirectory verifica que exista la carpeta raíz en el sistema de archivos
// Si no existe o está corrupta, la recrea
func ensureRootDirectory(path string, partitionStart int32) error {
	// Intentar formatear el superbloque solo para recrear la carpeta raíz
	// NOTA: Esta es una solución temporal que recrea la estructura básica
	sb := ext2.SuperBlock{}
	err := sb.DeserializeSuperBlock(path, partitionStart)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Verificar si la carpeta raíz existe usando el inodo #0
	rootInode := &ext2.INode{}
	err = rootInode.Deserialize(path, int64(sb.SInodeStart))
	if err != nil || rootInode.IType[0] != '0' { // '0' es tipo directorio
		// La carpeta raíz no existe o está corrupta, hay que recrearla
		fmt.Println("Recreando carpeta raíz durante la recuperación...")

		// Recrear estructuras básicas del sistema de archivos (carpeta raíz)
		// Esta acción es similar a lo que hace CreateUsersFile pero solo crea el inodo raíz
		err = recreateRootDirectory(&sb, path)
		if err != nil {
			return fmt.Errorf("error al recrear la carpeta raíz: %v", err)
		}
	}

	return nil
}

// recreateRootDirectory recrea la carpeta raíz (/) en el sistema de archivos
// Esta función es crítica para la recuperación cuando el inodo raíz está corrupto
func recreateRootDirectory(sb *ext2.SuperBlock, path string) error {
	// Crear el inodo raíz (inodo #0)
	rootInode := &ext2.INode{
		IUid:   1,
		IGid:   1,
		ISize:  0,
		IAtime: float32(time.Now().Unix()),
		ICtime: float32(time.Now().Unix()),
		IMtime: float32(time.Now().Unix()),
		IBlock: [15]int32{0, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Primer bloque es 0
		IType:  [1]byte{'0'},                                                         // Tipo directorio
		IPerm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz en la posición inicial de la tabla de inodos
	err := rootInode.Serialize(path, int64(sb.SInodeStart))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos para marcar el primer inodo como usado
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marcar como usado el primer inodo en el bitmap
	_, err = file.Seek(int64(sb.SBmInodeStart), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Crear el bloque para la carpeta raíz (bloque #0)
	rootBlock := &ext2.DirBlock{
		BContent: [4]ext2.DirContent{
			{BName: [12]byte{'.'}, BInodo: 0},                                         // Referencia a sí mismo
			{BName: [12]byte{'.', '.'}, BInodo: 0},                                    // Referencia a padre (es el mismo)
			{BName: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, BInodo: 1}, // Apuntará al inodo de users.txt
			{BName: [12]byte{'-'}, BInodo: -1},                                        // Entrada libre
		},
	}

	// Serializar el bloque de directorio raíz en la posición inicial de la tabla de bloques
	err = rootBlock.Serialize(path, int64(sb.SBlockStart))
	if err != nil {
		return err
	}

	// Marcar como usado el primer bloque en el bitmap
	_, err = file.Seek(int64(sb.SBmBlockStart), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	fmt.Println("Carpeta raíz (/) recreada exitosamente durante la recuperación")
	return nil
}

// ensureUsersFile verifica que exista el archivo users.txt
// Si no existe o está corrupto, lo recrea con valores por defecto
func ensureUsersFile(path string, partitionStart int32) error {
	// Implementar verificación de users.txt y recreación si es necesario
	// Esta función es similar a la anterior pero para el archivo users.txt
	sb := ext2.SuperBlock{}
	err := sb.DeserializeSuperBlock(path, partitionStart)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Intentar leer el archivo users.txt
	content, err := sb.ReadFile(path, []string{}, "users.txt")
	if err != nil || !strings.Contains(content, "root") {
		// El archivo users.txt no existe o está corrupto, hay que recrearlo
		fmt.Println("Recreando archivo users.txt durante la recuperación...")

		// Contenido por defecto para users.txt
		usersText := "1,G,root\n1,U,root,root,123\n"

		// Recrear el archivo users.txt
		err = recreateUsersFile(&sb, path, usersText)
		if err != nil {
			return fmt.Errorf("error al recrear users.txt: %v", err)
		}
	}

	return nil
}

// recreateUsersFile recrea el archivo users.txt con el contenido básico
func recreateUsersFile(sb *ext2.SuperBlock, path string, usersText string) error {
	// Crear el inodo para users.txt (inodo #1)
	usersInode := &ext2.INode{
		IUid:   1,
		IGid:   1,
		ISize:  int32(len(usersText)),
		IAtime: float32(time.Now().Unix()),
		ICtime: float32(time.Now().Unix()),
		IMtime: float32(time.Now().Unix()),
		IBlock: [15]int32{1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Apunta al bloque #1
		IType:  [1]byte{'1'},                                                         // Tipo archivo
		IPerm:  [3]byte{'6', '6', '4'},                                               // Permisos rw-rw-r--
	}

	// Serializar el inodo de users.txt
	err := usersInode.Serialize(path, int64(sb.SInodeStart+sb.SInodeS))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos para marcar el segundo inodo como usado
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marcar como usado el segundo inodo en el bitmap
	_, err = file.Seek(int64(sb.SBmInodeStart+1), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Crear el bloque para el archivo users.txt
	usersBlock := &ext2.FileBlock{
		BContent: [64]byte{},
	}
	copy(usersBlock.BContent[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.SBlockStart+sb.SBlockS))
	if err != nil {
		return err
	}

	// Marcar como usado el segundo bloque en el bitmap
	_, err = file.Seek(int64(sb.SBmBlockStart+1), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	fmt.Println("Archivo users.txt recreado exitosamente durante la recuperación")
	return nil
}
