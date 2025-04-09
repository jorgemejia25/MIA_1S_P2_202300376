package reports

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func LSReport(
	path_file string,
	output_path string,
	id string,
) error {

	// Obtener partición e información
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		return fmt.Errorf("error al obtener partición: %v", err)
	}

	// Separar ruta en directorios padres y nombre de archivo
	parentDirs, fileName := utils.GetParentDirectories(path_file)

	// Leer superbloque
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)

	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	dirInodeIndex, err := superBlock.FindFileInode(partitionPath, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al encontrar inodo de '%s': %v", path_file, err)
	}

	dirInode := &ext2.INode{}
	err = dirInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(dirInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo: %v", err)
	}

	if dirInode.IType[0] != '0' {
		fmt.Printf("'%s' no es una carpeta\n", path_file)
		return nil
	}

	var dotBuilder strings.Builder
	dotBuilder.WriteString("digraph LSReport {\n")
	dotBuilder.WriteString("    bgcolor=\"#f5f6fa\";\n") // Fondo del gráfico
	dotBuilder.WriteString("    fontname=\"Arial\";\n")
	dotBuilder.WriteString("    node [shape=plaintext];\n")
	dotBuilder.WriteString("    InfoTable [label=<\n")
	dotBuilder.WriteString("    <table border='0' cellborder='1' cellspacing='0' cellpadding='6' color='#2c3e50'>\n") // Estilo de tabla

	// Título del reporte
	dotBuilder.WriteString("        <tr><td colspan='7' bgcolor='#2c3e50' align='center'><font color='white'><b>LS Report</b></font></td></tr>\n")

	// Encabezados con estilo
	dotBuilder.WriteString("        <tr bgcolor='#2c3e50'>\n")
	dotBuilder.WriteString("            <td><font >Permisos</font></td>\n")
	dotBuilder.WriteString("            <td><font >UID</font></td>\n")
	dotBuilder.WriteString("            <td><font >Inodo</font></td>\n")
	dotBuilder.WriteString("            <td><font >Size</font></td>\n")
	dotBuilder.WriteString("            <td><font >Fecha</font></td>\n")
	dotBuilder.WriteString("            <td><font >Tipo</font></td>\n")
	dotBuilder.WriteString("            <td><font >Nombre</font></td>\n")
	dotBuilder.WriteString("        </tr>\n")

	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			break
		}
		dirBlock := &ext2.DirBlock{}
		if err := dirBlock.Deserialize(partitionPath, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS))); err != nil {
			return fmt.Errorf("error al leer bloque %d: %v", blockIndex, err)
		}
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}
			name := strings.TrimRight(string(entry.BName[:]), "\x00")
			if name != "." && name != ".." {
				entryInodeIndex := entry.BInodo
				entryInode := &ext2.INode{}
				err := entryInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(entryInodeIndex*superBlock.SInodeS)))
				if err != nil {
					return fmt.Errorf("error al leer inodo del entry %d: %v", entryInodeIndex, err)
				}
				ctime := time.Unix(int64(entryInode.ICtime), 0).Format("2006-01-02 15:04:05")
				fileType := "Archivo"
				if entryInode.IType[0] == '0' {
					fileType = "Carpeta"
				}
				perms := string(entryInode.IPerm[:])
				fmt.Printf(
					"Permisos: %s, UID: %d, ID: %d, Size: %d, Fecha: %s, Tipo: %s, Nombre: %s\n",
					perms, entryInode.IUid, entryInodeIndex, entryInode.ISize, ctime, fileType, name,
				)
				dotBuilder.WriteString(fmt.Sprintf(
					"<tr bgcolor='#f9f9f9' align='left'><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
					perms, entryInode.IUid, entryInodeIndex, entryInode.ISize, ctime, fileType, name,
				))

			}
		}
	}

	dotBuilder.WriteString("</table>>];\n")
	dotBuilder.WriteString("}\n")

	f, err := os.Create(output_path + ".dot")
	if err != nil {
		return fmt.Errorf("error al crear archivo dot: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(dotBuilder.String()); err != nil {
		return fmt.Errorf("error al escribir archivo dot: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error al cerrar archivo dot: %v", err)
	}

	cmd := exec.Command("dot", "-Tpng", output_path+".dot", "-o", output_path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al generar PNG: %v", err)
	}

	return nil
}
