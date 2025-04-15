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

func SuperBlockReport(outputPath string, id string) error {
	partitionData, diskPath, err := memory.GetInstance().GetMountedPartition(id)

	fmt.Println("Generating SuperBlock report at:", outputPath)

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

	// Leer el superbloque desde la partición
	sb := &ext2.SuperBlock{}
	err = sb.DeserializeSuperBlock(diskPath, partitionData.Partition.Part_start)
	if err != nil {
		return err
	}

	// Convertir el tiempo de montaje y desmontaje a una fecha legible
	mountTime := time.Unix(int64(sb.SMtime), 0)
	unmountTime := time.Unix(int64(sb.SUmTime), 0)

	// Contenido del dot con estilos mejorados
	dotContent := fmt.Sprintf(`digraph G {
        bgcolor="#f7f7f7";
        node [shape=plaintext fontname="Arial" fontsize=12];
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="white" style="rounded">
                <tr><td colspan="2" bgcolor="#4b6584" color="white"><b>REPORTE DE SUPERBLOQUE</b></td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Sistema de archivos</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Cantidad de inodos</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Cantidad de bloques</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Inodos libres</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Bloques libres</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Tiempo de montaje</b></td><td>%s</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Tiempo de desmontaje</b></td><td>%s</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Contador de montajes</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Valor Magic</b></td><td>0x%X</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Tamaño de inodo</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Tamaño de bloque</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Primer inodo libre</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Primer bloque libre</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Inicio bitmap inodos</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Inicio bitmap bloques</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Inicio tabla inodos</b></td><td>%d</td></tr>
                <tr><td bgcolor="#ecf0f1"><b>Inicio tabla bloques</b></td><td>%d</td></tr>
            `,
		sb.SFilesystemType,
		sb.SInodesCount,
		sb.SBlocksCount,
		sb.SFreeInodesCount,
		sb.SFreeBlocksCount,
		mountTime.Format(time.RFC3339),
		unmountTime.Format(time.RFC3339),
		sb.SMntCount,
		sb.SMagic,
		sb.SInodeS,
		sb.SBlockS,
		sb.SFirstIno,
		sb.SFirstBlo,
		sb.SBmInodeStart,
		sb.SBmBlockStart,
		sb.SInodeStart,
		sb.SBlockStart)

	// Cierre de la tabla con estilo
	dotContent += `</table>> style="filled" fillcolor="white" color="#34495e" penwidth=2];
	
	// Añadimos un título
    label = "Reporte de SuperBlock";
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

	fmt.Println("SuperBlock report created successfully at:", outputPath)

	return nil
}
