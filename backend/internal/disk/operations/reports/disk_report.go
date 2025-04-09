package reports

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/utils"
)

type DiskPartition struct {
	Type       string // "MBR", "Primaria", "Extendida", "Lógica", "EBR", "Libre"
	Start      int64
	Size       int64
	Percentage float64
	Name       string
}

func DiskReport(outputPath string, id string) error {
	_, diskPath, err := memory.GetInstance().GetMountedPartition(id)

	fmt.Println("Generando reporte de disco en:", outputPath)

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

	// Deserializar el MBR
	mbr := structures.MBR{}
	err = mbr.DeserializeMBR(diskPath)
	if err != nil {
		return err
	}

	// Obtener el tamaño total del disco
	diskSize := mbr.Mbr_size

	// Obtener información del archivo de disco
	_, err = os.Stat(diskPath)
	if err != nil {
		return err
	}

	// Usar el nombre del archivo como nombre del disco
	diskName := filepath.Base(diskPath)

	// Lista para almacenar todas las particiones y espacios libres
	var partitions []DiskPartition

	// Agregar el MBR al inicio
	mbrSize := int64(structures.MBRSize)
	mbrPercentage := float64(mbrSize) / float64(diskSize) * 100
	partitions = append(partitions, DiskPartition{
		Type:       "MBR",
		Start:      0,
		Size:       mbrSize,
		Percentage: mbrPercentage,
		Name:       "MBR",
	})

	// Recopilar información de todas las particiones en el MBR
	var validPartitions []structures.Partition
	for _, p := range mbr.Mbr_partitions {
		if p.Part_status == '1' {
			validPartitions = append(validPartitions, p)
		}
	}

	// Ordenar particiones por su posición de inicio
	sort.Slice(validPartitions, func(i, j int) bool {
		return validPartitions[i].Part_start < validPartitions[j].Part_start
	})

	// Inicialización del punto de inicio para detectar espacios libres
	currentPos := mbrSize

	// Analizar cada partición y los espacios entre ellas
	for i, partition := range validPartitions {
		// Si hay espacio libre antes de esta partición
		if int64(partition.Part_start) > currentPos {
			freeSize := int64(partition.Part_start) - currentPos
			freePercentage := float64(freeSize) / float64(diskSize) * 100

			partitions = append(partitions, DiskPartition{
				Type:       "Libre",
				Start:      currentPos,
				Size:       freeSize,
				Percentage: freePercentage,
				Name:       fmt.Sprintf("Espacio libre %d", i+1),
			})
		}

		// Agregar la partición actual
		partType := "Primaria"
		if partition.Part_type == 'E' {
			partType = "Extendida"
		}

		partPercentage := float64(partition.Part_size) / float64(diskSize) * 100
		// Limpiar el nombre de la partición
		partName := cleanName(string(partition.Part_name[:]))

		partitions = append(partitions, DiskPartition{
			Type:       partType,
			Start:      int64(partition.Part_start),
			Size:       int64(partition.Part_size),
			Percentage: partPercentage,
			Name:       partName,
		})

		// Si es una partición extendida, agregar particiones lógicas
		if partition.Part_type == 'E' {
			currentEBRStart := partition.Part_start

			for currentEBRStart != -1 {
				ebr := structures.EBR{}
				err := ebr.DeserializeEBR(diskPath, currentEBRStart)
				if err != nil {
					return err
				}

				// Agregar el EBR
				ebrSize := int64(structures.EBRSize)
				ebrPercentage := float64(ebrSize) / float64(diskSize) * 100

				partitions = append(partitions, DiskPartition{
					Type:       "EBR",
					Start:      int64(currentEBRStart),
					Size:       ebrSize,
					Percentage: ebrPercentage,
					Name:       "EBR",
				})

				// Si la partición lógica está activa
				if ebr.Part_mount == '1' && ebr.Part_size > 0 {
					logicalStart := int64(currentEBRStart) + ebrSize
					logicalPercentage := float64(ebr.Part_size) / float64(diskSize) * 100
					// Limpiar el nombre de la partición lógica
					logicalName := cleanName(string(ebr.Part_name[:]))

					partitions = append(partitions, DiskPartition{
						Type:       "Lógica",
						Start:      logicalStart,
						Size:       int64(ebr.Part_size),
						Percentage: logicalPercentage,
						Name:       logicalName,
					})
				}

				// Mover al siguiente EBR
				currentEBRStart = ebr.Part_next
			}
		}

		// Actualizar la posición actual
		currentPos = int64(partition.Part_start) + int64(partition.Part_size)
	}

	// Verificar si hay espacio libre al final del disco
	if currentPos < int64(diskSize) {
		freeSize := int64(diskSize) - currentPos
		freePercentage := float64(freeSize) / float64(diskSize) * 100

		partitions = append(partitions, DiskPartition{
			Type:       "Libre",
			Start:      currentPos,
			Size:       freeSize,
			Percentage: freePercentage,
			Name:       "Espacio libre final",
		})
	}

	// Generar contenido dot con un formato extremadamente simple para evitar problemas
	dotContent := `digraph G {
    bgcolor = "#f7f7f7";
    
    node [shape = plaintext; fontname = "Arial"; style = "filled"; fillcolor = "#FFFFFF"; color = "#333333"];
    
    edge [color = "#666666"; penwidth = 1.5];
    
    label = "Reporte DISK - ` + html.EscapeString(diskName) + `";
    labelloc = "t";
    fontsize = "20";
    fontname = "Arial Bold";
    fontcolor = "#2c3e50";
    
    disk [label = <<table border="0" cellborder="1" cellspacing="0" cellpadding="6" style="rounded">
        <tr>
            <td colspan="` + fmt.Sprintf("%d", len(partitions)) + `" bgcolor="#4b6584"><font color="white"><b>Estructura del Disco</b></font></td>
        </tr>
        <tr>`

	// Añadir las particiones con una estructura simplificada
	for _, p := range partitions {
		bgColor := "#FFFFFF" // Color por defecto
		textColor := "#000000"

		switch p.Type {
		case "MBR":
			bgColor = "#3498db" // Azul
			textColor = "#FFFFFF"
		case "Primaria":
			bgColor = "#2ecc71" // Verde
		case "Extendida":
			bgColor = "#e74c3c" // Rojo
		case "Lógica":
			bgColor = "#9b59b6" // Púrpura
		case "EBR":
			bgColor = "#f39c12" // Naranja
		case "Libre":
			bgColor = "#ecf0f1" // Gris claro
		}

		// Calcular el ancho de la celda (mínimo 1%)
		width := int(p.Percentage)
		if width < 1 {
			width = 1
		}

		// Escapar nombre para HTML y asegurar que sea seguro
		escapedName := html.EscapeString(p.Name)

		// Estructura simplificada sin tablas anidadas
		cellContent := fmt.Sprintf(`<td bgcolor="%s" width="%d"><font color="%s"><b>%s</b><br/>%.1f%%<br/>%s</font></td>`,
			bgColor, width, textColor, p.Type, p.Percentage, escapedName)

		dotContent += cellContent
	}

	dotContent += `
        </tr>
    </table>>];
}`

	// Escribir el archivo dot
	err = os.WriteFile(dotFileName, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo dot: %v", err)
	}

	// Convertir dot a imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Imprimir el contenido del archivo dot para depuración
		fmt.Println("Contenido del archivo DOT que causó el error:")
		fmt.Println(dotContent)
		return fmt.Errorf("error al generar la imagen con dot: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("Reporte de disco creado exitosamente en:", outputPath)

	return nil
}

// cleanName limpia el nombre de una partición eliminando caracteres nulos
// y otros caracteres problemáticos
func cleanName(name string) string {
	// Eliminar caracteres nulos
	name = strings.ReplaceAll(name, "\x00", "")

	// Trim espacios
	name = strings.TrimSpace(name)

	// También podemos eliminar bytes nulos usando bytes.Trim
	nameBytes := []byte(name)
	nameBytes = bytes.Trim(nameBytes, "\x00")

	return string(nameBytes)
}
