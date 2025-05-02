package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"disk.simulator.com/m/v2/internal/commands"

	"github.com/gin-gonic/gin"
)

type CommandRequest struct {
	Command string `json:"command"`
}

func HandleCommand(c *gin.Context) {
	var req CommandRequest

	// Hacer bind del JSON al struct
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"output": err.Error(),
		})
		return
	}

	// Separar el comando línea por línea
	lines := strings.Split(req.Command, "\n")
	var output []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignorar líneas vacías
		if line == "" {
			continue
		}

		// Si la línea es un comentario completo, ignorarla
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Eliminar cualquier comentario que esté después del comando
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Si después de quitar el comentario la línea está vacía, ignorarla
		if line == "" {
			continue
		}

		// Get the command prefix
		parts := strings.Split(line, " ")
		command := strings.ToLower(parts[0])

		// Convertir el comando en la línea a minúsculas para mantener consistencia
		lowercaseLine := strings.Replace(line, parts[0], command, 1)

		// Verificar qué tipo de comando es
		isDiskCommand := isDiskCommand(command)
		isPartitionCommand := isPartitionCommand(command)

		var cmdOutput string
		var err error

		if isDiskCommand {
			// Ejecutar como comando de disco
			cmdOutput, err = commands.ParseDiskCommand(command, lowercaseLine)
		} else if isPartitionCommand {
			// Ejecutar como comando de partición
			cmdOutput, err = commands.ParsePartitionCommand(command, lowercaseLine)
		} else if isAuthCommand(command) {
			// Ejecutar como comando de autenticación
			cmdOutput, err = commands.ParseAuthCommand(command, lowercaseLine)
		} else {
			// Comando desconocido
			err = fmt.Errorf("comando desconocido: %s", command)
		}

		if err != nil {
			output = append(output, fmt.Sprintf("Error: %s \n", err.Error()))
		} else {
			println(cmdOutput)
			output = append(output, cmdOutput)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"output": strings.Join(output, "\n"),
	})
}

// isDiskCommand verifica si el comando es un comando de disco
func isDiskCommand(cmd string) bool {
	diskCommands := []string{"mkdisk", "rmdisk", "fdisk", "rep", "mount", "mounted", "unmount", "journaling", "recovery", "loss"}
	return containsIgnoreCase(diskCommands, cmd)
}

// isPartitionCommand verifica si el comando es un comando de partición
func isPartitionCommand(cmd string) bool {
	partitionCommands := []string{"mkfs", "mkdir", "mkfile", "cat", "rename", "move", "copy", "find", "chown", "chmod", "rm", "edit", "ls", "df", "du", "stat", "cp", "mv", "remove"}
	return containsIgnoreCase(partitionCommands, cmd)
}

// isAuthCommand verifica si el comando es un comando de autenticación
func isAuthCommand(cmd string) bool {
	authCommands := []string{"login", "logout", "mkgrp", "mkusr", "rmgrp", "rmusr", "chgrp"}
	return containsIgnoreCase(authCommands, cmd)
}

// containsIgnoreCase verifica si una cadena está en un slice ignorando mayúsculas y minúsculas
func containsIgnoreCase(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {

			return true
		}
	}
	return false
}
