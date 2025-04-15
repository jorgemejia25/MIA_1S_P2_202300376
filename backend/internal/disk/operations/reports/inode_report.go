package reports

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func InodeReport(
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
        label="Reporte de Inodos"
        labelloc="t"
        fontsize="20"
        fontname="Arial"
        rankdir=LR
    `

	for i := int32(0); i < superBlock.SInodesCount; i++ {
		inode := ext2.INode{}

		err = inode.Deserialize(path, int64(superBlock.SInodeStart+(i*superBlock.SInodeS)))

		if err != nil {
			return err
		}

		// Determinar el color según el tipo de inodo (directorio vs archivo)
		nodeColor := "#D6EAF8"   // Azul claro para archivos
		headerColor := "#2874A6" // Azul oscuro para encabezados de archivos
		typeLabel := "ARCHIVO"

		if inode.IType[0] == '0' {
			nodeColor = "#D5F5E3"   // Verde claro para directorios
			headerColor = "#1E8449" // Verde oscuro para encabezados de directorios
			typeLabel = "DIRECTORIO"
		}

		atime := time.Unix(int64(inode.IAtime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.ICtime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.IMtime), 0).Format(time.RFC3339)

		// Crear un tooltip con información sobre el inodo
		tooltip := fmt.Sprintf("Inodo %d: %s, Permisos: %s, Tamaño: %d bytes",
			i, typeLabel, string(inode.IPerm[:]), inode.ISize)

		dotContent += fmt.Sprintf(`inode%d [tooltip="%s", label=<
            <table border="0" cellborder="1" cellspacing="0" cellpadding="4" style="rounded" bgcolor="%s">
                <tr><td colspan="2" bgcolor="%s" align="center"><font color="white"><b> INODO %d - %s </b></font></td></tr>
                <tr><td bgcolor="#F8F9F9"><b>UID</b></td><td>%d</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>GID</b></td><td>%d</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Tamaño</b></td><td>%d bytes</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Acceso</b></td><td>%s</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Creación</b></td><td>%s</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Modificación</b></td><td>%s</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Tipo</b></td><td>%s (%c)</td></tr>
                <tr><td bgcolor="#F8F9F9"><b>Permisos</b></td><td>%s</td></tr>
                <tr><td colspan="2" bgcolor="%s" align="center"><font color="white"><b>BLOQUES DIRECTOS</b></font></td></tr>
            `, i, tooltip, nodeColor, headerColor, i, typeLabel, inode.IUid, inode.IGid, inode.ISize,
			atime, ctime, mtime, typeLabel, rune(inode.IType[0]), string(inode.IPerm[:]), headerColor)

		// Bloques directos con estilo
		for j, block := range inode.IBlock {
			if j > 11 {
				break
			}
			bgColor := "#F8F9F9" // Color de fondo para bloques sin usar
			cellValue := fmt.Sprintf("%d", block)

			if block == -1 {
				cellValue = "No usado"
				bgColor = "#F2F3F4" // Gris muy claro para bloques no usados
			} else {
				bgColor = "#E8F8F5" // Verde muy claro para bloques usados
			}

			dotContent += fmt.Sprintf(`<tr><td bgcolor="%s"><b>Bloque %d</b></td><td>%s</td></tr>`,
				bgColor, j+1, cellValue)
		}

		// Bloques indirectos con estilo
		dotContent += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="%s" align="center"><font color="white"><b>BLOQUES INDIRECTOS</b></font></td></tr>
			`, headerColor)

		// Indirecto simple
		bgColor := "#F8F9F9"
		cellValue := fmt.Sprintf("%d", inode.IBlock[12])
		if inode.IBlock[12] == -1 {
			cellValue = "No usado"
			bgColor = "#F2F3F4"
		} else {
			bgColor = "#E8F8F5"
		}
		dotContent += fmt.Sprintf(`<tr><td bgcolor="%s"><b>Indirecto Simple</b></td><td>%s</td></tr>`,
			bgColor, cellValue)

		// Indirecto doble
		bgColor = "#F8F9F9"
		cellValue = fmt.Sprintf("%d", inode.IBlock[13])
		if inode.IBlock[13] == -1 {
			cellValue = "No usado"
			bgColor = "#F2F3F4"
		} else {
			bgColor = "#E8F8F5"
		}
		dotContent += fmt.Sprintf(`<tr><td bgcolor="%s"><b>Indirecto Doble</b></td><td>%s</td></tr>`,
			bgColor, cellValue)

		// Indirecto triple
		bgColor = "#F8F9F9"
		cellValue = fmt.Sprintf("%d", inode.IBlock[14])
		if inode.IBlock[14] == -1 {
			cellValue = "No usado"
			bgColor = "#F2F3F4"
		} else {
			bgColor = "#E8F8F5"
		}
		dotContent += fmt.Sprintf(`<tr><td bgcolor="%s"><b>Indirecto Triple</b></td><td>%s</td></tr>`,
			bgColor, cellValue)

		dotContent += `</table>>];
		`

		// Agregar enlace al siguiente inodo si no es el último
		if i < superBlock.SInodesCount-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d [weight=2];\n", i, i+1)
		}
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

	fmt.Printf("Reporte de inodos generado en %s\n", outputPath)

	return nil
}
