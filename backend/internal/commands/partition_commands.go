package commands

import (
	"fmt"
	"strings"

	"disk.simulator.com/m/v2/internal/args"

	partition_operations "disk.simulator.com/m/v2/internal/disk/operations/partitions"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var partitionRootCmd = &cobra.Command{Use: "partition"} // Renombrar rootCmd a partitionRootCmd

var mkfsCmd = &cobra.Command{
	Use:   "mkfs",
	Short: "Format a partition",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		fsType, _ := cmd.Flags().GetString("type")

		if id == "" {
			return fmt.Errorf("el id es requerido")
		}

		if fsType == "" {
			fsType = "full" // Valor predeterminado
		}

		// Normalizar el tipo para comparaciones
		fsType = strings.ToLower(fsType)

		// Crear el output formateado
		output := fmt.Sprintf("Formatting partition %s with filesystem type %s", id, fsType)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Aquí iría la lógica para formatear la partición
		err := partition_operations.FormatPartition(id, fsType)

		if err != nil {
			return err
		}

		return nil
	},
}

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "Create a directory in a partition",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		p, _ := cmd.Flags().GetBool("p")

		if p {
			fmt.Println("Flag -p es true")
		} else {
			fmt.Println("Flag -p es false")
		}

		if path == "" {
			return fmt.Errorf("el path es requerido")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Creating directory in partition %s", path)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Aquí iría la lógica para crear el directorio
		err := partition_operations.CreateDirectory(path, p)
		if err != nil {
			return err
		}

		return nil
	},
}

var mkfileCmd = &cobra.Command{
	Use:   "mkfile",
	Short: "Create a file in a partition",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get size flag
		size, _ := cmd.Flags().GetInt("size")

		// Get path flag
		path, _ := cmd.Flags().GetString("path")

		if path == "" {
			return fmt.Errorf("path is required")
		}

		// Verificar que size no sea negativo
		if size < 0 {
			return fmt.Errorf("el tamaño (size) no puede ser negativo")
		}

		// Get content flag
		content, _ := cmd.Flags().GetString("cont")

		fmt.Println("Content:", content)
		fmt.Println("Size:", size)

		// Eliminamos la validación que exige size>0 o content
		// Si ambos están vacíos, se creará un archivo vacío con size=0

		// Get r bool flag
		r, _ := cmd.Flags().GetBool("r")

		// Create the formatted output
		output := fmt.Sprintf("Creating file in partition %s", path)

		// Write the output to the command output
		fmt.Fprintln(cmd.OutOrStdout(), output)

		err := partition_operations.CreateFile(path, size, content, r)

		if err != nil {
			return err
		}

		return nil
	},
}

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "Display content of one or more files",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Obtener todos los flags desde el comando
		flags := cmd.Flags()
		var fileContents []string
		filesFound := false

		// Buscar todos los flags y procesar los que empiezan con "file"
		flags.VisitAll(func(flag *pflag.Flag) {
			if strings.HasPrefix(flag.Name, "file") && flag.Changed {
				filesFound = true
				filePath := flag.Value.String()
				if filePath != "" {
					content, err := partition_operations.CatFile(filePath)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "Error leyendo archivo %s: %v\n", filePath, err)
					} else {
						fileContents = append(fileContents,
							fmt.Sprintf("=== %s ===\n%s", filePath, content))
					}
				}
			}
		})

		if !filesFound {
			return fmt.Errorf("debe especificar al menos un archivo (ej: --file1=ruta)")
		}

		// Mostrar contenido de cada archivo
		output := strings.Join(fileContents, "\n\n")
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

func init() {
	// MKFS
	partitionRootCmd.AddCommand(mkfsCmd)
	mkfsCmd.PersistentFlags().StringP("id", "i", "", "ID of the partition") // Agregar alias -i para --id
	mkfsCmd.MarkPersistentFlagRequired("id")
	mkfsCmd.Flags().StringP("type", "t", "full", "Filesystem type (ext4, ntfs, etc.)") // Agregar alias -t para --type

	// MKDIR
	partitionRootCmd.AddCommand(mkdirCmd)
	mkdirCmd.PersistentFlags().StringP("path", "a", "", "Path of the directory") // Agregar alias -p para --path
	mkdirCmd.MarkPersistentFlagRequired("path")
	mkdirCmd.Flags().BoolP("p", "p", false, "Create parent directories automatically") // Agregar alias -p para --p

	// MKFILE
	partitionRootCmd.AddCommand(mkfileCmd)
	mkfileCmd.PersistentFlags().StringP("path", "a", "", "Path of the file") // Agregar alias -a para --path
	mkfileCmd.MarkPersistentFlagRequired("path")
	mkfileCmd.PersistentFlags().IntP("size", "s", 0, "Size of the file in bytes (opcional si se proporciona content)")
	mkfileCmd.PersistentFlags().StringP("cont", "c", "", "Content of the file or path to file using @/path/to/file format (opcional si se proporciona size)")
	mkfileCmd.Flags().BoolP("r", "r", false, "Create parent directories automatically")

	// CAT
	partitionRootCmd.AddCommand(catCmd)

	// Predefinimos algunos flags comunes (se pueden agregar más si es necesario)
	for i := 1; i <= 10; i++ {
		catCmd.Flags().String(fmt.Sprintf("file%d", i), "", fmt.Sprintf("Ruta al archivo %d", i))
	}
}

// ParsePartitionCommand analiza y ejecuta un comando de partición
func ParsePartitionCommand(command string, data string) (string, error) {
	// Divide los argumentos respetando las comillas y los flags con valores unidos por "="
	args := args.SplitArgs(data)

	// Reiniciar los flags a sus valores predeterminados antes de ejecutar un nuevo comando
	resetPartitionFlags()

	// Configura los argumentos para cobra
	partitionRootCmd.SetArgs(args)

	// Captura la salida del comando
	output := &strings.Builder{}
	partitionRootCmd.SetOut(output)

	// Ejecuta el comando
	err := partitionRootCmd.Execute()
	if err != nil {
		return "", err
	}

	// Validar argumentos desconocidos
	if len(partitionRootCmd.Flags().Args()) > 0 {
		return "", fmt.Errorf("unknown arguments: %v", partitionRootCmd.Flags().Args())
	}

	// Devolver la salida capturada
	return output.String(), nil
}

// resetPartitionFlags reinicia los valores de todos los flags a sus valores predeterminados
func resetPartitionFlags() {
	// Reiniciar flags de mkfs
	if mkfsCmd.Flags().Lookup("id") != nil {
		mkfsCmd.Flags().Set("id", "")
	}
	if mkfsCmd.Flags().Lookup("type") != nil {
		mkfsCmd.Flags().Set("type", "full")
	}

	// Reiniciar flags de mkdir
	if mkdirCmd.Flags().Lookup("path") != nil {
		mkdirCmd.Flags().Set("path", "")
	}
	if mkdirCmd.Flags().Lookup("p") != nil {
		mkdirCmd.Flags().Set("p", "false")
	}

	// Reiniciar flags de mkfile
	if mkfileCmd.Flags().Lookup("path") != nil {
		mkfileCmd.Flags().Set("path", "")
	}
	if mkfileCmd.Flags().Lookup("size") != nil {
		mkfileCmd.Flags().Set("size", "0")
	}
	if mkfileCmd.Flags().Lookup("cont") != nil {
		mkfileCmd.Flags().Set("cont", "")
	}
	if mkfileCmd.Flags().Lookup("r") != nil {
		mkfileCmd.Flags().Set("r", "false")
	}

	// Reiniciar flags de cat (reiniciar todos los posibles file1...file10)
	for i := 1; i <= 10; i++ {
		flagName := fmt.Sprintf("file%d", i)
		if catCmd.Flags().Lookup(flagName) != nil {
			catCmd.Flags().Set(flagName, "")
		}
	}
}
