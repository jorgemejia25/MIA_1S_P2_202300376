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
		fs, _ := cmd.Flags().GetString("fs")

		ext3 := true // Cambiado a true por defecto (ext3)

		// Solo cambia a false si explícitamente se indica "2fs" (ext2)
		if fs == "2fs" {
			ext3 = false
		}

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
		err := partition_operations.FormatPartition(id, fsType, ext3)

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

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")

		if path == "" {
			return fmt.Errorf("path is required")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Removing file or directory in partition %s", path)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		err := partition_operations.RemoveFileOrDirectory(path)

		if err != nil {
			return err
		}

		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		contenido, _ := cmd.Flags().GetString("contenido")

		if path == "" {
			return fmt.Errorf("path is required")
		}

		if contenido == "" {
			return fmt.Errorf("contenido is required")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Editing file in partition %s", path)

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		err := partition_operations.EditFile(path, contenido)

		if err != nil {
			return err
		}

		return nil
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename a file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		oldPath, _ := cmd.Flags().GetString("path")
		newName, _ := cmd.Flags().GetString("name")

		if oldPath == "" || newName == "" {
			return fmt.Errorf("se requieren tanto el path como el nuevo nombre")
		}

		output := fmt.Sprintf("Renombrando %s a %s", oldPath, newName)
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return partition_operations.RenameFile(oldPath, newName)
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy a file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("path")
		dest, _ := cmd.Flags().GetString("destino")

		if source == "" || dest == "" {
			return fmt.Errorf("se requieren tanto el source como el dest")
		}

		output := fmt.Sprintf("Copiando %s a %s", source, dest)
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return partition_operations.CopyFileOrDirectory(source, dest)
	},
}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("path")
		dest, _ := cmd.Flags().GetString("destino")

		if source == "" || dest == "" {
			return fmt.Errorf("se requieren tanto el source como el dest")
		}

		output := fmt.Sprintf("Moviendo %s a %s", source, dest)
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return partition_operations.MoveFileOrDirectory(source, dest)
	},
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")

		if path == "" || name == "" {
			return fmt.Errorf("se requieren tanto el path como el nombre")
		}

		output, err := partition_operations.FindFileOrFolderTree(path, name)

		if err != nil {
			return fmt.Errorf("error al buscar: %v", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var chownCmd = &cobra.Command{
	Use:   "chown",
	Short: "Cambiar propietario de un archivo o directorio",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		usuario, _ := cmd.Flags().GetString("usuario")
		r, _ := cmd.Flags().GetBool("r")

		if path == "" {
			return fmt.Errorf("error: se requiere la ruta del archivo o directorio (--path)")
		}

		if usuario == "" {
			return fmt.Errorf("error: se requiere el nombre de usuario (--usuario)")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Cambiando propietario de %s al usuario %s", path, usuario)
		if r {
			output += " (recursivamente)"
		}

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return partition_operations.ChangeOwner(path, usuario, r)
	},
}

var chmodCmd = &cobra.Command{
	Use:   "chmod",
	Short: "Cambiar permisos de un archivo o directorio",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		ugo, _ := cmd.Flags().GetString("ugo")
		r, _ := cmd.Flags().GetBool("r")

		if path == "" {
			return fmt.Errorf("error: se requiere la ruta del archivo o directorio (--path)")
		}

		if ugo == "" {
			return fmt.Errorf("error: se requieren los permisos en formato [0-7][0-7][0-7] (--ugo)")
		}

		// Crear el output formateado
		output := fmt.Sprintf("Cambiando permisos de %s a %s", path, ugo)
		if r {
			output += " (recursivamente)"
		}

		// Escribir el output en la salida del comando
		fmt.Fprintln(cmd.OutOrStdout(), output)

		return partition_operations.ChangePermissions(path, ugo, r)
	},
}

func init() {
	// MKFS
	partitionRootCmd.AddCommand(mkfsCmd)
	mkfsCmd.PersistentFlags().StringP("id", "i", "", "ID of the partition") // Agregar alias -i para --id
	mkfsCmd.MarkPersistentFlagRequired("id")
	mkfsCmd.Flags().StringP("type", "t", "full", "Filesystem type (ext4, ntfs, etc.)") // Agregar alias -t para --type
	mkfsCmd.Flags().StringP("fs", "e", "", "Use ext3 filesystem")                      // Agregar alias -e para --ext3
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

	// REMOVE
	partitionRootCmd.AddCommand(removeCmd)
	removeCmd.PersistentFlags().StringP("path", "p", "", "Path of the file or directory to remove") // Agregar alias -p para --path
	removeCmd.MarkPersistentFlagRequired("path")

	// EDIT
	partitionRootCmd.AddCommand(editCmd)
	editCmd.PersistentFlags().StringP("path", "p", "", "Path of the file or directory to edit") // Agregar alias -p para --path
	editCmd.MarkPersistentFlagRequired("path")
	editCmd.PersistentFlags().StringP("contenido", "c", "", "Content of the file") // Agregar alias -c para --content
	editCmd.MarkPersistentFlagRequired("contenido")
	// CAT
	partitionRootCmd.AddCommand(catCmd)

	// Predefinimos algunos flags comunes (se pueden agregar más si es necesario)
	for i := 1; i <= 10; i++ {
		catCmd.Flags().String(fmt.Sprintf("file%d", i), "", fmt.Sprintf("Ruta al archivo %d", i))
	}

	// RENAME
	partitionRootCmd.AddCommand(renameCmd)
	renameCmd.PersistentFlags().StringP("path", "p", "", "Ruta actual del archivo/directorio")
	renameCmd.PersistentFlags().StringP("name", "n", "", "Nuevo nombre")
	renameCmd.MarkPersistentFlagRequired("path")
	renameCmd.MarkPersistentFlagRequired("name")

	// COPY
	partitionRootCmd.AddCommand(copyCmd)
	copyCmd.PersistentFlags().StringP("path", "s", "", "Ruta origen")
	copyCmd.PersistentFlags().StringP("destino", "d", "", "Ruta destino")
	copyCmd.MarkPersistentFlagRequired("path")
	copyCmd.MarkPersistentFlagRequired("destino")

	// MOVE
	partitionRootCmd.AddCommand(moveCmd)
	moveCmd.PersistentFlags().StringP("path", "s", "", "Ruta origen")
	moveCmd.PersistentFlags().StringP("destino", "d", "", "Ruta destino")
	moveCmd.MarkPersistentFlagRequired("path")
	moveCmd.MarkPersistentFlagRequired("destino")

	// FIND
	partitionRootCmd.AddCommand(findCmd)
	findCmd.PersistentFlags().StringP("path", "p", "", "Ruta origen")
	findCmd.PersistentFlags().StringP("name", "n", "", "Nombre del archivo/directorio a buscar")
	findCmd.MarkPersistentFlagRequired("path")
	findCmd.MarkPersistentFlagRequired("name")

	// CHOWN
	partitionRootCmd.AddCommand(chownCmd)
	chownCmd.PersistentFlags().StringP("path", "p", "", "Ruta del archivo/directorio")
	chownCmd.PersistentFlags().StringP("usuario", "u", "", "Nombre del usuario")
	chownCmd.PersistentFlags().BoolP("r", "r", false, "Cambiar propietario recursivamente")
	chownCmd.MarkPersistentFlagRequired("path")
	chownCmd.MarkPersistentFlagRequired("usuario")

	// CHMOD
	partitionRootCmd.AddCommand(chmodCmd)
	chmodCmd.PersistentFlags().StringP("path", "p", "", "Ruta del archivo/directorio")
	chmodCmd.PersistentFlags().StringP("ugo", "u", "", "Permisos en formato [0-7][0-7][0-7]")
	chmodCmd.PersistentFlags().BoolP("r", "r", false, "Cambiar permisos recursivamente")
	chmodCmd.MarkPersistentFlagRequired("path")
	chmodCmd.MarkPersistentFlagRequired("ugo")
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

	// Reiniciar flags de copy
	if copyCmd.Flags().Lookup("path") != nil {
		copyCmd.Flags().Set("path", "")
	}
	if copyCmd.Flags().Lookup("destino") != nil {
		copyCmd.Flags().Set("destino", "")
	}

	// Reiniciar flags de move
	if moveCmd.Flags().Lookup("path") != nil {
		moveCmd.Flags().Set("path", "")
	}
	if moveCmd.Flags().Lookup("destino") != nil {
		moveCmd.Flags().Set("destino", "")
	}

	// Reiniciar flags de chown
	if chownCmd.Flags().Lookup("path") != nil {
		chownCmd.Flags().Set("path", "")
	}
	if chownCmd.Flags().Lookup("usuario") != nil {
		chownCmd.Flags().Set("usuario", "")
	}
	if chownCmd.Flags().Lookup("r") != nil {
		chownCmd.Flags().Set("r", "false")
	}

	// Reiniciar flags de chmod
	if chmodCmd.Flags().Lookup("path") != nil {
		chmodCmd.Flags().Set("path", "")
	}
	if chmodCmd.Flags().Lookup("ugo") != nil {
		chmodCmd.Flags().Set("ugo", "")
	}
	if chmodCmd.Flags().Lookup("r") != nil {
		chmodCmd.Flags().Set("r", "false")
	}
}
