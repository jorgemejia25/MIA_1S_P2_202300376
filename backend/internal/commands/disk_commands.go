package commands

import (
	"fmt"
	"strconv"
	"strings"

	"disk.simulator.com/m/v2/internal/args"
	disk_operations "disk.simulator.com/m/v2/internal/disk/operations/disk"
	partition_operations "disk.simulator.com/m/v2/internal/disk/operations/partitions"
	"disk.simulator.com/m/v2/internal/disk/operations/reports"
	"disk.simulator.com/m/v2/internal/disk/types"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "disk"}

var mkdiskCmd = &cobra.Command{
	Use:   "mkdisk",
	Short: "Create a new disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		fit, _ := cmd.Flags().GetString("fit")
		unit, _ := cmd.Flags().GetString("unit")
		size, _ := cmd.Flags().GetInt("size")

		// si unit es nil o vacío, asignar valor predeterminado
		if unit == "" {
			unit = "M"
		}

		// Convertir unit y fit a mayúsculas para mantener consistencia
		unit = strings.ToUpper(unit)
		fit = strings.ToUpper(fit)

		// Validar el valor de fit
		if fit != "WF" && fit != "FF" && fit != "BF" {
			return fmt.Errorf("invalid fit type. Use WF, FF, or BF")
		}

		// Validar el valor de unit
		if unit != "K" && unit != "M" {
			return fmt.Errorf("invalid unit type. Use K or M")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Creating disk at %s with size %d%s, fit %s", path, size, unit, fit)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Crear el disco usando el nuevo struct
		params := types.MkDisk{
			Path: path,
			Size: size,
			Unit: unit,
			Fit:  fit,
		}

		err := disk_operations.CreateDisk(params)
		if err != nil {
			return err
		}

		return nil
	},
}

var rmdiskCmd = &cobra.Command{
	Use:   "rmdisk",
	Short: "Remove an existing disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")

		// Crear el output formateado
		output := fmt.Sprintf("Removing disk at %s", path)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Eliminar el disco
		err := disk_operations.RemoveDisk(path)
		if err != nil {
			return err
		}

		return nil
	},
}

var fdiskCmd = &cobra.Command{
	Use:   "fdisk",
	Short: "Manage disk partitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		size, _ := cmd.Flags().GetInt("size")
		unit, _ := cmd.Flags().GetString("unit")
		fit, _ := cmd.Flags().GetString("fit")
		partitionType, _ := cmd.Flags().GetString("type")
		del, _ := cmd.Flags().GetString("delete")
		add, _ := cmd.Flags().GetString("add")

		// Convertir unit y fit a mayúsculas para mantener consistencia
		if unit == "" {
			unit = "M" // Unidad predeterminada
		} else {
			unit = strings.ToUpper(unit)
		}
		fit = strings.ToUpper(fit)
		partitionType = strings.ToUpper(partitionType)

		// Validar el valor de unit
		if unit != "B" && unit != "K" && unit != "M" {
			return fmt.Errorf("invalid unit type. Use B, K or M")
		}

		// Validar el valor de type solo si no estamos eliminando o modificando espacio
		if del == "" && add == "" && partitionType != "P" && partitionType != "E" && partitionType != "L" {
			return fmt.Errorf("invalid partition type. Use P, E, or L")
		}

		// Validar el valor de fit
		if fit == "" {
			fit = "FF" // Valor predeterminado
		}

		// Crear la estructura FDisk con los parámetros
		params := types.FDisk{
			Path: path,
			Size: size,
			Unit: unit,
			Fit:  fit,
			Name: name,
			Type: partitionType,
			Del:  del,
			Add: func() int {
				if add != "" {
					addInt, err := strconv.Atoi(add)
					if err == nil {
						return addInt
					}
				}
				return 0
			}(),
		}

		// Si es una operación de eliminación
		if del != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Deleting partition %s from %s\n", params.Name, params.Path)
			err := partition_operations.DeletePartition(params)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Partition deleted successfully")
			return nil
		}

		// Si es una operación de añadir o quitar espacio
		if add != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Modifying partition %s at %s by %d%s\n",
				params.Name, params.Path, params.Add, params.Unit)
			err := partition_operations.AddSpacePartition(params)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Space modified in partition successfully")
			return nil
		}

		// Si es una operación de creación
		fmt.Fprintf(cmd.OutOrStdout(), "Creating partition %s at %s with size %d%s, type %s\n",
			params.Name, params.Path, params.Size, params.Unit, params.Type)
		err := partition_operations.CreatePartition(params)
		if err != nil {
			return err
		}

		return nil
	},
}

var repCmd = &cobra.Command{
	Use:   "rep",
	Short: "Print the MBR of a disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		pathFileLs, _ := cmd.Flags().GetString("path_file_ls")

		fmt.Println("Generando reporte")
		// pathFileLs, _ := cmd.Flags().GetString("path_file_ls")

		if path == "" {
			return fmt.Errorf("el path es requerido")
		}

		if name == "" {
			return fmt.Errorf("el nombre es requerido")
		}

		if id == "" {
			return fmt.Errorf("el ID es requerido")
		}

		// Normalizar el nombre para hacer la comparación insensible a mayúsculas/minúsculas y espacios
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		reportProcessed := false

		if normalizedName == "mbr" {
			err := reports.MbrReport(path, id)
			if err != nil {
				return fmt.Errorf("error al imprimir el MBR: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte MBR generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "inode" {
			err := reports.InodeReport(path, id)
			if err != nil {
				return fmt.Errorf("error al imprimir el Inode: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte Inode generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "disk" {
			// fmt.Println("Generando reporte de Disco")

			err := reports.DiskReport(path, id)
			if err != nil {
				return fmt.Errorf("error al imprimir el Disco: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte Disco generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "bm_inode" {
			err := reports.BInodeReport(path, id)
			if err != nil {
				return fmt.Errorf("error al imprimir el Bitmap de Inode: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte Bitmap de Inode generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "bm_block" {
			err := reports.BBlockReport(path, id)
			if err != nil {
				return fmt.Errorf("error al imprimir el Bitmap de Bloque: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte Bitmap de Bloque generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "sb" {
			err := reports.SuperBlockReport(path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de SuperBlock: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte de SuperBlock generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "block" {
			err := reports.BlockReport(path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de Bloque: %v", err)
			}

			// Crear el output formateado
			output := fmt.Sprintf("Reporte de Bloque generado en %s", path)

			// Escribir el output en la salida del comando
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "file" {
			err := reports.FileReport(pathFileLs, path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de Archivo: %v", err)
			}

			output := fmt.Sprintf("Reporte de Archivo generado en %s", pathFileLs)
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "ls" {
			err := reports.LSReport(pathFileLs, path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de LS: %v", err)
			}

			output := fmt.Sprintf("Reporte de LS generado en %s", path)
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "tree" {
			err := reports.TreeReport(path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de Tree: %v", err)
			}

			output := fmt.Sprintf("Reporte de Tree generado en %s", path)
			fmt.Fprintln(cmd.OutOrStdout(), output)
			reportProcessed = true
		}

		if normalizedName == "journaling" {
			reportText, err := reports.JournalingReport(path, id)
			if err != nil {
				return fmt.Errorf("error al generar el reporte de Journaling: %v", err)
			}

			// Mostrar el reporte de journaling directamente en la consola
			fmt.Fprintln(cmd.OutOrStdout(), reportText)
			reportProcessed = true
		}

		// Solo mostrar el mensaje de error si no se procesó ningún reporte
		if !reportProcessed {
			output := fmt.Sprintf("Argumento name: %s desconocido", name)
			fmt.Fprintln(cmd.OutOrStdout(), output)
		}

		return nil
	},
}

var mountedCmd = &cobra.Command{
	Use:   "mounted",
	Short: "List all mounted partitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		output := partition_operations.GetMountedPartitions()
		fmt.Fprintln(cmd.OutOrStdout(), output)
		return nil
	},
}

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount a partition",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")

		// Crear el output formateado
		output := fmt.Sprintf("Mounting partition %s from disk at %s", name, path)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Montar la partición
		err := partition_operations.MountPartition(name, path)
		if err != nil {
			return err
		}

		return nil
	},
}

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount a partition",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")

		if id == "" {
			return fmt.Errorf("el ID de la partición es requerido")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Unmounting partition with ID %s", id)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Desmontar la partición
		err := partition_operations.UnmountPartition(id)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Partition unmounted successfully")
		return nil
	},
}

var journalingCmd = &cobra.Command{
	Use:   "journaling",
	Short: "Muestra información de todas las transacciones realizadas en una partición",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")

		if id == "" {
			return fmt.Errorf("el ID es requerido")
		}

		// Generar el reporte de journaling directamente
		reportText, err := reports.JournalingReport("", id)
		if err != nil {
			return fmt.Errorf("error al generar el reporte de Journaling: %v", err)
		}

		// Mostrar el reporte de journaling directamente en la consola
		fmt.Fprintln(cmd.OutOrStdout(), reportText)

		return nil
	},
}

var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Recupera archivos y carpetas desde el journaling",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")

		if id == "" {
			return fmt.Errorf("el ID es requerido")
		}

		// Ejecutar la recuperación desde el journaling
		output, err := partition_operations.RecoverFromJournaling(id)
		if err != nil {
			return fmt.Errorf("error en la recuperación: %v", err)
		}

		// Imprimir la salida formateada
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var lossCmd = &cobra.Command{
	Use:   "loss",
	Short: "Simula una pérdida de información en el sistema de archivos",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")

		if id == "" {
			return fmt.Errorf("el ID es requerido")
		}

		// Ejecutar la simulación de pérdida
		output, err := partition_operations.SimulateSystemLoss(id)
		if err != nil {
			return fmt.Errorf("error en la simulación de pérdida: %v", err)
		}

		// Imprimir la salida formateada
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var disklistCmd = &cobra.Command{
	Use:   "disklist",
	Short: "Listar todos los discos creados",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Obtener información de los discos
		disksInfo, err := disk_operations.GetDisksInfo()
		if err != nil {
			return err
		}

		// Imprimir la información de los discos
		fmt.Fprintln(cmd.OutOrStdout(), disksInfo)
		return nil
	},
}

var partlistCmd = &cobra.Command{
	Use:   "partlist",
	Short: "Listar todas las particiones de un disco",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")

		if path == "" {
			return fmt.Errorf("el parámetro path es requerido")
		}

		// Obtener información de las particiones
		partitionsInfo, err := partition_operations.GetPartitionsInfo(path)
		if err != nil {
			return err
		}

		// Imprimir la información de las particiones
		fmt.Fprintln(cmd.OutOrStdout(), partitionsInfo)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdiskCmd)
	rootCmd.AddCommand(rmdiskCmd)
	rootCmd.AddCommand(repCmd)
	rootCmd.AddCommand(fdiskCmd)
	rootCmd.AddCommand(mountedCmd)
	rootCmd.AddCommand(mountCmd)
	rootCmd.AddCommand(unmountCmd)
	rootCmd.AddCommand(journalingCmd)
	rootCmd.AddCommand(recoveryCmd)
	rootCmd.AddCommand(lossCmd)
	rootCmd.AddCommand(disklistCmd)
	rootCmd.AddCommand(partlistCmd)

	// MKDISK
	mkdiskCmd.PersistentFlags().IntP("size", "s", 0, "Size of the disk in MB or KB") // Agregar alias -s para --size
	mkdiskCmd.MarkPersistentFlagRequired("size")

	mkdiskCmd.PersistentFlags().StringP("path", "p", "", "Path to the disk") // Agregar alias -p para --path
	mkdiskCmd.MarkPersistentFlagRequired("path")

	mkdiskCmd.Flags().StringP("fit", "f", "FF", "Fit type (WF, FF, BF)") // Agregar alias -f para --fit
	mkdiskCmd.Flags().StringP("unit", "u", "M", "Unit type (K, M)")      // Agregar alias -u para --unit

	// RMDISK
	rmdiskCmd.PersistentFlags().StringP("path", "p", "", "Path to the disk") // Agregar alias -p para --path
	rmdiskCmd.MarkPersistentFlagRequired("path")

	// REP
	repCmd.PersistentFlags().StringP("path", "p", "", "Path to the disk") // Agregar alias -p para --path
	repCmd.MarkPersistentFlagRequired("path")

	repCmd.PersistentFlags().StringP("name", "n", "", "Name of the partition") // Agregar alias -n para --name
	repCmd.MarkPersistentFlagRequired("name")

	repCmd.PersistentFlags().String("id", "", "ID of the partition") // Agregar alias -i para --id
	repCmd.MarkPersistentFlagRequired("id")

	repCmd.Flags().String("path_file_ls", "", "Path to the file to save the ls command") // Agregar alias -f para --path_file_ls

	// FDISK
	fdiskCmd.PersistentFlags().StringP("path", "p", "", "Path to the disk") // Agregar alias -p para --path
	fdiskCmd.MarkPersistentFlagRequired("path")

	fdiskCmd.PersistentFlags().StringP("name", "n", "", "Name of the partition") // Agregar alias -n para --name
	fdiskCmd.MarkPersistentFlagRequired("name")

	fdiskCmd.PersistentFlags().IntP("size", "s", 0, "Size of the partition in MB or KB") // Agregar alias -s para --size
	fdiskCmd.MarkPersistentFlagRequired("size")

	fdiskCmd.Flags().StringP("unit", "u", "M", "Unit type (B, K, M)")              // Unidad predeterminada actualizada a M
	fdiskCmd.Flags().StringP("fit", "f", "FF", "Fit type (WF, FF, BF)")            // Agregar alias -f para --fit
	fdiskCmd.Flags().StringP("type", "t", "P", "Type of the partition (P, E, L)")  // Agregar alias -t para --type
	fdiskCmd.Flags().StringP("delete", "d", "", "Delete partition (Full or Fast)") // Agregar alias -d para --delete
	fdiskCmd.Flags().StringP("add", "a", "", "ADD to the partition")               // Agregar alias -a para --start

	// MOUNT
	mountCmd.PersistentFlags().StringP("path", "p", "", "Path to the disk")
	mountCmd.MarkPersistentFlagRequired("path")
	mountCmd.PersistentFlags().StringP("name", "n", "", "Name of the partition to mount")
	mountCmd.MarkPersistentFlagRequired("name")

	// UNMOUNT
	unmountCmd.PersistentFlags().StringP("id", "i", "", "ID of the partition to unmount")
	unmountCmd.MarkPersistentFlagRequired("id")

	// JOURNALING
	journalingCmd.PersistentFlags().StringP("id", "i", "", "ID de la partición para mostrar el journaling")
	journalingCmd.MarkPersistentFlagRequired("id")

	// RECOVERY
	recoveryCmd.PersistentFlags().StringP("id", "i", "", "ID de la partición para recuperar archivos desde el journaling")
	recoveryCmd.MarkPersistentFlagRequired("id")

	// LOSS
	lossCmd.PersistentFlags().StringP("id", "i", "", "ID de la partición para simular pérdida de sistema")
	lossCmd.MarkPersistentFlagRequired("id")

	// PARTLIST
	partlistCmd.PersistentFlags().StringP("path", "p", "", "Ruta del disco")
	partlistCmd.MarkPersistentFlagRequired("path")
}

// ParseDiskCommand analiza y ejecuta un comando de disco
func ParseDiskCommand(command string, data string) (string, error) {
	// Divide los argumentos respetando las comillas y los flags con valores unidos por "="
	args := args.SplitArgs(data)

	// Reiniciar los flags a sus valores predeterminados antes de ejecutar un nuevo comando
	// Esto evita que los valores del comando anterior se mantengan
	resetFlags()

	// Configura los argumentos para cobra
	rootCmd.SetArgs(args)

	// Captura la salida del comando
	output := &strings.Builder{}
	rootCmd.SetOut(output)

	// Ejecuta el comando
	err := rootCmd.Execute()
	if err != nil {
		return "", err
	}

	// Validar argumentos desconocidos
	if len(rootCmd.Flags().Args()) > 0 {
		return "", fmt.Errorf("unknown arguments: %v", rootCmd.Flags().Args())
	}

	// Devolver la salida capturada
	return output.String(), nil
}

// resetFlags reinicia los valores de todos los flags a sus valores predeterminados
func resetFlags() {
	// Reiniciar flags de mkdisk
	if mkdiskCmd.Flags().Lookup("size") != nil {
		mkdiskCmd.Flags().Set("size", "0")
	}
	if mkdiskCmd.Flags().Lookup("path") != nil {
		mkdiskCmd.Flags().Set("path", "")
	}
	if mkdiskCmd.Flags().Lookup("fit") != nil {
		mkdiskCmd.Flags().Set("fit", "FF")
	}
	if mkdiskCmd.Flags().Lookup("unit") != nil {
		mkdiskCmd.Flags().Set("unit", "M")
	}

	// Reiniciar flags de rmdisk
	if rmdiskCmd.Flags().Lookup("path") != nil {
		rmdiskCmd.Flags().Set("path", "")
	}

	// Reiniciar flags de fdisk
	if fdiskCmd.Flags().Lookup("path") != nil {
		fdiskCmd.Flags().Set("path", "")
	}
	if fdiskCmd.Flags().Lookup("name") != nil {
		fdiskCmd.Flags().Set("name", "")
	}
	if fdiskCmd.Flags().Lookup("size") != nil {
		fdiskCmd.Flags().Set("size", "0")
	}
	if fdiskCmd.Flags().Lookup("unit") != nil {
		fdiskCmd.Flags().Set("unit", "M")
	}
	if fdiskCmd.Flags().Lookup("fit") != nil {
		fdiskCmd.Flags().Set("fit", "FF")
	}
	if fdiskCmd.Flags().Lookup("type") != nil {
		fdiskCmd.Flags().Set("type", "P")
	}

	// Reiniciar flags de rep
	if repCmd.Flags().Lookup("path") != nil {
		repCmd.Flags().Set("path", "")
	}
	if repCmd.Flags().Lookup("name") != nil {
		repCmd.Flags().Set("name", "")
	}
	if repCmd.Flags().Lookup("id") != nil {
		repCmd.Flags().Set("id", "")
	}
	if repCmd.Flags().Lookup("path_file_ls") != nil {
		repCmd.Flags().Set("path_file_ls", "")
	}

	// Reiniciar flags de mount
	if mountCmd.Flags().Lookup("path") != nil {
		mountCmd.Flags().Set("path", "")
	}
	if mountCmd.Flags().Lookup("name") != nil {
		mountCmd.Flags().Set("name", "")
	}

	// Reiniciar flags de unmount
	if unmountCmd.Flags().Lookup("id") != nil {
		unmountCmd.Flags().Set("id", "")
	}

	// Reiniciar flags de journaling
	if journalingCmd.Flags().Lookup("id") != nil {
		journalingCmd.Flags().Set("id", "")
	}

	// Reiniciar flags de recovery
	if recoveryCmd.Flags().Lookup("id") != nil {
		recoveryCmd.Flags().Set("id", "")
	}

	// Reiniciar flags de loss
	if lossCmd.Flags().Lookup("id") != nil {
		lossCmd.Flags().Set("id", "")
	}

	// Reiniciar flags de partlist
	if partlistCmd.Flags().Lookup("path") != nil {
		partlistCmd.Flags().Set("path", "")
	}
}
