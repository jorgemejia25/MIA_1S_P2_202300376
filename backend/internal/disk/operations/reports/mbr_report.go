package reports

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/utils"
)

func MbrReport(outputPath string, id string) error {
	_, diskPath, err := memory.GetInstance().GetMountedPartition(id)

	fmt.Println("Generating MBR report at:", outputPath)

	if err != nil {
		return err
	}

	// Crear las carpetas necesarias para el archivo de salida
	err = utils.CreateParentDirs(outputPath)
	if err != nil {
		return err
	}

	// Generar el nombre para el archivo .dot basado en la ruta de salida
	dotFileName := filepath.Join(filepath.Dir(outputPath), filepath.Base(outputPath)+".dot")

	mbr := structures.MBR{}

	err = mbr.DeserializeMBR(diskPath)
	if err != nil {
		return err
	}

	fmt.Println("MBR size:", mbr.Mbr_size)

	// Contenido del dot con estilos mejorados
	dotContent := fmt.Sprintf(`digraph G {
        bgcolor="#f7f7f7";
        node [shape=plaintext fontname="Arial" fontsize=12];
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="white" style="rounded">
                <tr><td colspan="2" bgcolor="#4b6584" color="white"><b>REPORTE DE MBR</b></td></tr>
                <tr><td bgcolor="#ecf0f1"><b>mbr_tamano</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>mrb_fecha_creacion</b></td><td>%s</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>mbr_disk_signature</b></td><td>%d</td></tr>
            `, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	for i, partition := range mbr.Mbr_partitions {
		// Convertimos los caracteres a string para evitar problemas de formato
		statusStr := string(partition.Part_status)
		typeStr := string(partition.Part_type)
		fitStr := string(partition.Part_fit)

		// Limpiamos el nombre para asegurar que no contenga caracteres nulos
		nameStr := strings.ReplaceAll(string(partition.Part_name[:]), "\x00", "")

		// Color de fondo para el encabezado de partición según tipo
		headerColor := "#3498db" // Azul por defecto
		if partition.Part_type == 'P' {
			headerColor = "#2ecc71" // Verde para primarias
		} else if partition.Part_type == 'E' {
			headerColor = "#e74c3c" // Rojo para extendidas
		}

		dotContent += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="%s" color="white"><b>PARTICION %d</b></td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_status</b></td><td>%s</td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_type</b></td><td>%s</td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_fit</b></td><td>%s</td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_start</b></td><td>%d</td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_size</b></td><td>%d</td></tr>
			<tr><td bgcolor="#ecf0f1"><b>part_name</b></td><td>%s</td></tr>
			`, headerColor, i, statusStr, typeStr, fitStr, partition.Part_start, partition.Part_size, nameStr)

		if partition.Part_type == 'E' {
			currentEBRStart := partition.Part_start
			logicalPartitionNumber := 1

			for currentEBRStart != -1 {
				ebr := structures.EBR{}
				err := ebr.DeserializeEBR(diskPath, currentEBRStart)
				if err != nil {
					return err
				}

				// Convertir caracteres a strings para evitar problemas
				ebrStatusStr := string(ebr.Part_mount)
				ebrFitStr := string(ebr.Part_fit)
				ebrNameStr := strings.ReplaceAll(string(bytes.Trim(ebr.Part_name[:], "\x00")), "\x00", "")

				dotContent += fmt.Sprintf(`
					<tr><td colspan="2" bgcolor="#9b59b6" color="white"><b>Partición lógica %d</b></td></tr>
					<tr><td bgcolor="#f0e6f6"><b>part_status</b></td><td>%s</td></tr>
					<tr><td bgcolor="#f0e6f6"><b>part_fit</b></td><td>%s</td></tr>
					<tr><td bgcolor="#f0e6f6"><b>part_start</b></td><td>%d</td></tr>
					<tr><td bgcolor="#f0e6f6"><b>part_size</b></td><td>%d</td></tr>
					<tr><td bgcolor="#f0e6f6"><b>part_name</b></td><td>%s</td></tr>
					`, logicalPartitionNumber, ebrStatusStr, ebrFitStr, ebr.Part_start, ebr.Part_size, ebrNameStr)

				// Mover al siguiente EBR
				currentEBRStart = ebr.Part_next
				logicalPartitionNumber++
			}
		}
	}

	// Cierre de la tabla con estilo
	dotContent += `</table>> style="filled" fillcolor="white" color="#34495e" penwidth=2];
	
	// Añadimos un título
    label = "Reporte de Master Boot Record (MBR)";
    labelloc = "t";
    fontname = "Arial Bold";
    fontsize = 20;
    fontcolor = "#2c3e50";
	}`

	// Crear el archivo .dot
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

	fmt.Println("MBR report created successfully at:", outputPath)

	return nil
}
