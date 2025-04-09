package reports

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

// generateInodeNodeDOT crea el código DOT para representar un inodo con todos sus detalles
func generateInodeNodeDOT(inode *ext2.INode, inodeIndex int32, name string) string {
	// Determinar el color según el tipo de inodo
	nodeColor := "#D5F5E3"   // Verde claro para directorios
	headerColor := "#1E8449" // Verde oscuro para encabezados de directorios
	typeLabel := "DIRECTORIO"

	if inode.IType[0] == '1' {
		nodeColor = "#D6EAF8"   // Azul para archivos
		headerColor = "#2874A6" // Azul oscuro para encabezados de archivos
		typeLabel = "ARCHIVO"
	}

	// Tooltip con información del inodo
	tooltip := fmt.Sprintf("Inodo %d: %s, Permisos: %s, Tamaño: %d bytes",
		inodeIndex, typeLabel, string(inode.IPerm[:]), inode.ISize)

	// Iniciar la definición del nodo
	nodeDef := fmt.Sprintf(`inode%d [tooltip="%s", label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4" style="rounded" bgcolor="%s">
			<tr><td colspan="2" bgcolor="%s" align="center"><font color="white"><b>INODO %d - %s</b></font></td></tr>
			<tr><td bgcolor="#F8F9F9"><b>Nombre</b></td><td>%s</td></tr>
			<tr><td bgcolor="#F8F9F9"><b>Tamaño</b></td><td>%d bytes</td></tr>
			<tr><td bgcolor="#F8F9F9"><b>Tipo</b></td><td>%s (%c)</td></tr>
			<tr><td bgcolor="#F8F9F9"><b>Permisos</b></td><td>%s</td></tr>
			<tr><td colspan="2" bgcolor="%s" align="center"><font color="white"><b>BLOQUES DIRECTOS</b></font></td></tr>
		`, inodeIndex, tooltip, nodeColor, headerColor, inodeIndex, typeLabel, name,
		inode.ISize,
		typeLabel, rune(inode.IType[0]), string(inode.IPerm[:]), headerColor)

	// Agregar los bloques directos
	for j, block := range inode.IBlock {
		if j > 11 {
			break
		}
		// Solo mostrar los bloques que están en uso (diferentes de -1)
		if block != -1 {
			bgColor := "#E8F8F5" // Verde muy claro para bloques usados
			nodeDef += fmt.Sprintf(`<tr><td bgcolor="%s"><b>Bloque %d</b></td><td>%d</td></tr>`,
				bgColor, j+1, block)
		}
	}

	// Cerrar la tabla y el nodo
	nodeDef += `</table>>];`

	return nodeDef
}

// generateBlockNodeDOT crea el código DOT para representar cualquier tipo de bloque
func generateBlockNodeDOT(block interface{}, blockIndex int32) string {
	switch typedBlock := block.(type) {
	case *ext2.DirBlock:
		// Bloque de directorio (naranja)
		blockDef := fmt.Sprintf(`block%d [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#FDEBD0">
				<tr><td colspan="2" bgcolor="#E67E22" align="center"><font color="white"><b>Bloque %d (Directorio)</b></font></td></tr>
				<tr><td bgcolor="#F8F9F9"><b>Nombre</b></td><td bgcolor="#F8F9F9"><b>Inodo</b></td></tr>`, blockIndex, blockIndex)

		// Mostrar las entradas del directorio
		for _, content := range typedBlock.BContent {
			if content.BInodo != -1 {
				name := strings.Trim(string(content.BName[:]), "\x00")
				if name == "" {
					name = "-"
				}
				blockDef += fmt.Sprintf(`<tr><td>%s</td><td>%d</td></tr>`, name, content.BInodo)
			}
		}

		blockDef += `</table>>];`
		return blockDef

	case *ext2.FileBlock:
		// Bloque de archivo (azul)
		content := strings.Trim(string(typedBlock.BContent[:]), "\x00")
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		// Escapar caracteres especiales para DOT
		content = strings.Replace(content, `"`, `\"`, -1)
		content = strings.Replace(content, "<", "&lt;", -1)
		content = strings.Replace(content, ">", "&gt;", -1)

		return fmt.Sprintf(`block%d [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#D6EAF8">
				<tr><td bgcolor="#2874A6" align="center"><font color="white"><b>Bloque %d (Archivo)</b></font></td></tr>
				<tr><td>%s</td></tr>
			</table>
		>];`, blockIndex, blockIndex, content)

	case *ext2.PointerBlock:
		// Bloque de punteros (celeste)
		blockDef := fmt.Sprintf(`block%d [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#A9CCE3">
				<tr><td colspan="2" bgcolor="#5DADE2" align="center"><font color="white"><b>Bloque %d (Punteros)</b></font></td></tr>
				<tr><td bgcolor="#F8F9F9"><b>Índice</b></td><td bgcolor="#F8F9F9"><b>Bloque</b></td></tr>
		`, blockIndex, blockIndex)

		// Mostrar los punteros válidos
		for i, ptr := range typedBlock.PContent {
			if ptr != -1 {
				blockDef += fmt.Sprintf(`<tr><td>%d</td><td>%d</td></tr>`, i, ptr)
			}
		}

		blockDef += `</table>>];`
		return blockDef

	default:
		// En caso de un tipo de bloque desconocido
		return fmt.Sprintf(`block%d [label="Bloque %d (Desconocido)"];`, blockIndex, blockIndex)
	}
}

// Función recursiva para procesar inodos y bloques
func processInode(
	superBlock *ext2.SuperBlock,
	path string,
	inodeIndex int32,
	name string,
	nodeDefinitions *[]string,
	nodeConnections *[]string,
	processedInodes map[int32]bool,
) error {
	// Verificar que el inodo esté dentro del rango válido
	if inodeIndex < 0 || inodeIndex >= superBlock.SInodesCount {
		return fmt.Errorf("número de inodo %d fuera de rango (0-%d)", inodeIndex, superBlock.SInodesCount-1)
	}

	// Evitar procesar inodos ya visitados
	if processedInodes[inodeIndex] {
		return nil
	}
	processedInodes[inodeIndex] = true

	// Obtener el inodo actual
	inode, err := superBlock.GetInodeByNumber(path, inodeIndex)
	if err != nil {
		return fmt.Errorf("error al obtener inodo %d: %v", inodeIndex, err)
	}

	// Generar el nodo DOT para el inodo actual
	inodeDef := generateInodeNodeDOT(inode, inodeIndex, name)
	*nodeDefinitions = append(*nodeDefinitions, inodeDef)

	// Procesar los bloques directos del inodo (0-11)
	for j := 0; j < 12; j++ {
		block := inode.IBlock[j]
		if block == -1 {
			continue
		}

		// Conectar el inodo con su bloque
		*nodeConnections = append(*nodeConnections, fmt.Sprintf("inode%d -> block%d;", inodeIndex, block))

		// Obtener el bloque según el tipo de inodo
		var blockDef string

		if inode.IType[0] == '0' { // Si es un directorio
			// Obtener como bloque de directorio
			dirBlock := &ext2.DirBlock{}
			err := dirBlock.Deserialize(path, int64(superBlock.SBlockStart+(block*superBlock.SBlockS)))
			if err == nil {
				blockDef = generateBlockNodeDOT(dirBlock, block)
				*nodeDefinitions = append(*nodeDefinitions, blockDef)

				// Procesar entradas de directorio
				for _, content := range dirBlock.BContent {
					if content.BInodo != -1 {
						// Verificar que el inodo hijo esté en un rango válido
						if content.BInodo < 0 || content.BInodo >= superBlock.SInodesCount {
							// Ignorar inodos inválidos
							continue
						}

						childName := strings.Trim(string(content.BName[:]), "\x00")
						if childName == "" {
							childName = "-"
						}

						// Solo procesar inodos que no sean "." o ".."
						if childName != "." && childName != ".." {
							// Conectar el bloque con el inodo hijo
							*nodeConnections = append(*nodeConnections,
								fmt.Sprintf("block%d -> inode%d;", block, content.BInodo))

							// Procesar recursivamente
							_ = processInode(superBlock, path, content.BInodo, childName,
								nodeDefinitions, nodeConnections, processedInodes)
						}
					}
				}
			}
		} else if inode.IType[0] == '1' { // Si es un archivo
			// Obtener como bloque de archivo
			fileBlock := &ext2.FileBlock{}
			err := fileBlock.Deserialize(path, int64(superBlock.SBlockStart+(block*superBlock.SBlockS)))
			if err == nil {
				blockDef = generateBlockNodeDOT(fileBlock, block)
				*nodeDefinitions = append(*nodeDefinitions, blockDef)
			}
		}
	}

	// Procesar bloque de punteros indirectos simples (posición 12)
	if inode.IBlock[12] != -1 {
		blockIndex := inode.IBlock[12]
		pointerBlock := &ext2.PointerBlock{}
		err := pointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS)))
		if err == nil {
			// Conectar el inodo con su bloque de punteros
			*nodeConnections = append(*nodeConnections, fmt.Sprintf("inode%d -> block%d [label=\"Simple\"];", inodeIndex, blockIndex))

			// Generar y agregar la definición del bloque de punteros
			blockDef := generateBlockNodeDOT(pointerBlock, blockIndex)
			*nodeDefinitions = append(*nodeDefinitions, blockDef)

			// Procesar cada puntero válido
			for i, ptr := range pointerBlock.PContent {
				if ptr != -1 {
					// Conectar el bloque de punteros con el bloque de datos
					*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", blockIndex, ptr, i))

					// Si el inodo es un archivo, procesar el bloque de datos
					if inode.IType[0] == '1' {
						fileBlock := &ext2.FileBlock{}
						err := fileBlock.Deserialize(path, int64(superBlock.SBlockStart+(ptr*superBlock.SBlockS)))
						if err == nil {
							dataBlockDef := generateBlockNodeDOT(fileBlock, ptr)
							*nodeDefinitions = append(*nodeDefinitions, dataBlockDef)
						}
					}
				}
			}
		}
	}

	// Procesar bloque de punteros indirectos dobles (posición 13)
	if inode.IBlock[13] != -1 {
		doubleBlockIndex := inode.IBlock[13]
		doublePointerBlock := &ext2.PointerBlock{}
		err := doublePointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(doubleBlockIndex*superBlock.SBlockS)))
		if err == nil {
			// Conectar el inodo con su bloque de punteros dobles
			*nodeConnections = append(*nodeConnections, fmt.Sprintf("inode%d -> block%d [label=\"Doble\"];", inodeIndex, doubleBlockIndex))

			// Generar y agregar la definición del bloque de punteros dobles
			blockDef := generateBlockNodeDOT(doublePointerBlock, doubleBlockIndex)
			*nodeDefinitions = append(*nodeDefinitions, blockDef)

			// Procesar cada puntero válido en el bloque doble
			for i, ptr := range doublePointerBlock.PContent {
				if ptr != -1 {
					// Conectar el bloque de punteros dobles con el bloque de punteros simples
					*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", doubleBlockIndex, ptr, i))

					// Procesar el bloque de punteros simples
					simplePointerBlock := &ext2.PointerBlock{}
					err := simplePointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(ptr*superBlock.SBlockS)))
					if err == nil {
						simpleBlockDef := generateBlockNodeDOT(simplePointerBlock, ptr)
						*nodeDefinitions = append(*nodeDefinitions, simpleBlockDef)

						// Procesar cada puntero válido en el bloque simple
						for j, dataPtr := range simplePointerBlock.PContent {
							if dataPtr != -1 {
								// Conectar el bloque de punteros simples con el bloque de datos
								*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", ptr, dataPtr, j))

								// Si el inodo es un archivo, procesar el bloque de datos
								if inode.IType[0] == '1' {
									fileBlock := &ext2.FileBlock{}
									err := fileBlock.Deserialize(path, int64(superBlock.SBlockStart+(dataPtr*superBlock.SBlockS)))
									if err == nil {
										dataBlockDef := generateBlockNodeDOT(fileBlock, dataPtr)
										*nodeDefinitions = append(*nodeDefinitions, dataBlockDef)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Procesar bloque de punteros indirectos triples (posición 14)
	if inode.IBlock[14] != -1 {
		tripleBlockIndex := inode.IBlock[14]
		triplePointerBlock := &ext2.PointerBlock{}
		err := triplePointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(tripleBlockIndex*superBlock.SBlockS)))
		if err == nil {
			// Conectar el inodo con su bloque de punteros triples
			*nodeConnections = append(*nodeConnections, fmt.Sprintf("inode%d -> block%d [label=\"Triple\"];", inodeIndex, tripleBlockIndex))

			// Generar y agregar la definición del bloque de punteros triples
			blockDef := generateBlockNodeDOT(triplePointerBlock, tripleBlockIndex)
			*nodeDefinitions = append(*nodeDefinitions, blockDef)

			// Procesar cada puntero válido en el bloque triple
			for i, ptr := range triplePointerBlock.PContent {
				if ptr != -1 {
					// Conectar el bloque de punteros triples con el bloque de punteros dobles
					*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", tripleBlockIndex, ptr, i))

					// Procesar el bloque de punteros dobles
					doublePointerBlock := &ext2.PointerBlock{}
					err := doublePointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(ptr*superBlock.SBlockS)))
					if err == nil {
						doubleBlockDef := generateBlockNodeDOT(doublePointerBlock, ptr)
						*nodeDefinitions = append(*nodeDefinitions, doubleBlockDef)

						// Procesar cada puntero válido en el bloque doble
						for j, simplePtr := range doublePointerBlock.PContent {
							if simplePtr != -1 {
								// Conectar el bloque de punteros dobles con el bloque de punteros simples
								*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", ptr, simplePtr, j))

								// Procesar el bloque de punteros simples
								simplePointerBlock := &ext2.PointerBlock{}
								err := simplePointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(simplePtr*superBlock.SBlockS)))
								if err == nil {
									simpleBlockDef := generateBlockNodeDOT(simplePointerBlock, simplePtr)
									*nodeDefinitions = append(*nodeDefinitions, simpleBlockDef)

									// Procesar cada puntero válido en el bloque simple
									for k, dataPtr := range simplePointerBlock.PContent {
										if dataPtr != -1 {
											// Conectar el bloque de punteros simples con el bloque de datos
											*nodeConnections = append(*nodeConnections, fmt.Sprintf("block%d -> block%d [label=\"%d\"];", simplePtr, dataPtr, k))

											// Si el inodo es un archivo, procesar el bloque de datos
											if inode.IType[0] == '1' {
												fileBlock := &ext2.FileBlock{}
												err := fileBlock.Deserialize(path, int64(superBlock.SBlockStart+(dataPtr*superBlock.SBlockS)))
												if err == nil {
													dataBlockDef := generateBlockNodeDOT(fileBlock, dataPtr)
													*nodeDefinitions = append(*nodeDefinitions, dataBlockDef)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func TreeReport(
	outputPath string,
	id string,
) error {

	partition, path, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		return err
	}

	superBlock := ext2.SuperBlock{}
	superBlock.DeserializeSuperBlock(path, partition.Partition.Part_start)
	superBlock.Print()

	err = utils.CreateParentDirs(outputPath)
	if err != nil {
		return err
	}

	dotFileName := filepath.Join(filepath.Dir(outputPath), filepath.Base(outputPath)+".dot")

	dotContent := `digraph G {
        bgcolor="#f7f7f7"
        node [shape=plaintext, fontname="Arial", style="filled", fillcolor="#FFFFFF", color="#333333"]
        edge [color="#666666", penwidth=1.5]
        label="Reporte de Árbol"
        labelloc="t"
        fontsize="20"
        fontname="Arial"
        rankdir=LR
    `

	// Preparar estructuras para recopilar nodos y conexiones
	var nodeDefinitions []string
	var nodeConnections []string
	processedInodes := make(map[int32]bool)

	// Procesar el inodo raíz recursivamente
	err = processInode(&superBlock, path, 0, "/", &nodeDefinitions, &nodeConnections, processedInodes)
	if err != nil {
		return err
	}

	// Generar el contenido final del archivo DOT
	dotContent += "\n    // Node definitions\n"
	for _, nodeDef := range nodeDefinitions {
		dotContent += "    " + nodeDef + "\n"
	}

	dotContent += "\n    // Node connections\n"
	for _, conn := range nodeConnections {
		dotContent += "    " + conn + "\n"
	}

	dotContent += "}"

	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear archivo .dot: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en archivo .dot: %v", err)
	}

	// Crear la imagen PNG usando el archivo .dot
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error al generar la imagen con dot: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("Reporte de árbol generado en %s\n", outputPath)

	return nil
}
