package reports

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func BlockReport(
	outputPath string,
	id string,
) error {
	partition, path, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return err
	}

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(path, partition.Partition.Part_start)
	if err != nil {
		return err
	}

	err = utils.CreateParentDirs(outputPath)
	if err != nil {
		return err
	}

	dotFileName := filepath.Join(filepath.Dir(outputPath), filepath.Base(outputPath)+".dot")

	dotContent := `digraph G {
        bgcolor="#f7f7f7"
        node [shape=plaintext, fontname="Arial", style="filled", fillcolor="#FFFFFF", color="#333333"]
        edge [color="#666666", penwidth=1.5]
        label="Reporte de Bloques"
        labelloc="t"
        fontsize="20"
        fontname="Arial"
        nodesep=0.5;
        ranksep=0.7;
        splines=false;
    `

	// Mapeo de bloques
	blockIdToInodeMap := make(map[int32]int32)
	validBlocks := []int32{}

	// Primero recorremos todos los inodos para mapear bloques a los inodos que los usan
	for i := int32(0); i < superBlock.SInodesCount; i++ {
		inode := ext2.INode{}
		err = inode.Deserialize(path, int64(superBlock.SInodeStart+(i*superBlock.SInodeS)))
		if err != nil {
			continue
		}

		// Para cada bloque directo del inodo
		for _, blockIdx := range inode.IBlock {
			if blockIdx != -1 {
				blockIdToInodeMap[blockIdx] = i
				validBlocks = append(validBlocks, blockIdx)
			}
		}
	}

	// Ahora procesamos cada bloque
	processedBlocks := make(map[int32]bool)

	// Número de bloques por fila para crear un grid
	const blocksPerRow = 4

	// Determinar cuántas filas necesitamos
	numRows := int(math.Ceil(float64(len(validBlocks)) / float64(blocksPerRow)))

	// Crear los subgrafos para las filas del grid
	for row := 0; row < numRows; row++ {
		dotContent += fmt.Sprintf("\n    subgraph row_%d {\n        rank=same;\n", row)

		// Insertar los bloques de esta fila
		start := row * blocksPerRow
		end := (row + 1) * blocksPerRow
		if end > len(validBlocks) {
			end = len(validBlocks)
		}

		for i := start; i < end; i++ {
			blockIdx := validBlocks[i]

			if _, exists := processedBlocks[blockIdx]; exists {
				continue
			}

			ownerInodeID, hasOwner := blockIdToInodeMap[blockIdx]
			if !hasOwner {
				continue // Saltamos bloques sin dueño
			}

			// Obtenemos información del inodo propietario
			ownerInode := ext2.INode{}
			err = ownerInode.Deserialize(path, int64(superBlock.SInodeStart+(ownerInodeID*superBlock.SInodeS)))
			if err != nil {
				continue
			}

			// Si el inodo es un directorio (tipo 0)
			if ownerInode.IType[0] == '0' {
				dirBlock := ext2.DirBlock{}
				err = dirBlock.Deserialize(path, int64(superBlock.SBlockStart+(blockIdx*superBlock.SBlockS)))
				if err != nil {
					continue
				}

				dotContent += fmt.Sprintf(`        block%d [tooltip="Bloque de Directorio %d", label=<
                    <table border="0" cellborder="1" cellspacing="0" cellpadding="4" style="rounded" bgcolor="#D5F5E3">
                        <tr><td colspan="2" bgcolor="#1E8449" align="center"><font color="white"><b>BLOQUE DE DIRECTORIO %d</b></font></td></tr>
                        <tr><td bgcolor="#F8F9F9"><b>Inodo Propietario</b></td><td>%d</td></tr>
                `, blockIdx, blockIdx, blockIdx, ownerInodeID)

				// Añadimos entradas de directorio
				for entryIdx, content := range dirBlock.BContent {
					if content.BInodo == -1 {
						continue
					}

					name := strings.TrimRight(string(content.BName[:]), "\x00")
					if name == "" {
						continue
					}

					dotContent += fmt.Sprintf(`
                        <tr><td colspan="2" bgcolor="#1E8449" align="center"><font color="white"><b>Entrada %d</b></font></td></tr>
                        <tr><td bgcolor="#F8F9F9"><b>Nombre</b></td><td>%s</td></tr>
                        <tr><td bgcolor="#F8F9F9"><b>Inodo</b></td><td>%d</td></tr>
                    `, entryIdx+1, name, content.BInodo)
				}

				dotContent += `</table>>];
                `
				processedBlocks[blockIdx] = true

			} else if ownerInode.IType[0] == '1' { // Si el inodo es un archivo (tipo 1)
				fileBlock := ext2.FileBlock{}
				err = fileBlock.Deserialize(path, int64(superBlock.SBlockStart+(blockIdx*superBlock.SBlockS)))
				if err != nil {
					continue
				}

				content := strings.TrimRight(string(fileBlock.BContent[:]), "\x00")
				if len(content) > 30 {
					content = content[:30] + "..."
				}

				dotContent += fmt.Sprintf(`        block%d [tooltip="Bloque de Archivo %d", label=<
                    <table border="0" cellborder="1" cellspacing="0" cellpadding="4" style="rounded" bgcolor="#D6EAF8">
                        <tr><td colspan="1" bgcolor="#2874A6" align="center"><font color="white"><b>BLOQUE DE ARCHIVO %d</b></font></td></tr>
                        <tr><td bgcolor="#F8F9F9"><b>Inodo Propietario</b></td></tr>
                        <tr><td>%d</td></tr>
                        <tr><td bgcolor="#2874A6" align="center"><font color="white"><b>Contenido</b></font></td></tr>
                        <tr><td>%s</td></tr>
                    </table>>];
                `, blockIdx, blockIdx, blockIdx, ownerInodeID, content)
				processedBlocks[blockIdx] = true

			} else if isPointerBlock(ownerInode, blockIdx) { // Si es un bloque de punteros
				pointerBlock := ext2.PointerBlock{}
				err = pointerBlock.Deserialize(path, int64(superBlock.SBlockStart+(blockIdx*superBlock.SBlockS)))
				if err != nil {
					continue
				}

				dotContent += fmt.Sprintf(`        block%d [tooltip="Bloque de Punteros %d", label=<
                    <table border="0" cellborder="1" cellspacing="0" cellpadding="4" style="rounded" bgcolor="#FADBD8">
                        <tr><td colspan="2" bgcolor="#943126" align="center"><font color="white"><b>BLOQUE DE PUNTEROS %d</b></font></td></tr>
                        <tr><td bgcolor="#F8F9F9"><b>Inodo Propietario</b></td><td>%d</td></tr>
                        <tr><td colspan="2" bgcolor="#943126" align="center"><font color="white"><b>Punteros a Bloques</b></font></td></tr>
                `, blockIdx, blockIdx, blockIdx, ownerInodeID)

				// Mostrar los punteros reales del bloque de punteros
				for ptr := 0; ptr < len(pointerBlock.PContent); ptr++ {
					if pointerBlock.PContent[ptr] != -1 {
						dotContent += fmt.Sprintf(`
                            <tr><td bgcolor="#F8F9F9"><b>Puntero %d</b></td><td>Bloque #%d</td></tr>
                            `, ptr+1, pointerBlock.PContent[ptr])
					}
				}

				dotContent += `</table>>];
                `
				processedBlocks[blockIdx] = true
			}
		}

		// Cerrar el subgrafo de esta fila
		dotContent += "    }\n"

		// Si hay una fila siguiente, conectar invisiblemente el último elemento de esta fila
		// con el primer elemento de la siguiente fila para mantener la estructura
		if row < numRows-1 && end < len(validBlocks) && start < end {
			nextRowStart := end
			if nextRowStart < len(validBlocks) {
				dotContent += fmt.Sprintf("    block%d -> block%d [style=invis];\n",
					validBlocks[end-1], validBlocks[nextRowStart])
			}
		}

		// Conectar los bloques en la misma fila con enlaces invisibles para mantener la distribución horizontal
		for i := start; i < end-1; i++ {
			dotContent += fmt.Sprintf("    block%d -> block%d [style=invis];\n",
				validBlocks[i], validBlocks[i+1])
		}
	}

	// Cerramos el gráfico
	dotContent += "}"

	// Escribir el archivo DOT
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

	fmt.Printf("Reporte de bloques generado en %s\n", outputPath)
	return nil
}

// Función auxiliar para determinar si un bloque es un bloque de punteros
func isPointerBlock(inode ext2.INode, blockIdx int32) bool {
	// Un bloque es de punteros si es referido por los bloques indirectos del inodo
	return inode.IBlock[12] == blockIdx || inode.IBlock[13] == blockIdx || inode.IBlock[14] == blockIdx
}
