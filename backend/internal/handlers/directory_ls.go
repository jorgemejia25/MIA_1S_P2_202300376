package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	partition_operations "disk.simulator.com/m/v2/internal/disk/operations/partitions"
	"github.com/gin-gonic/gin"
)

// Estructura para la respuesta del endpoint
type DirectoryLsResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

// HandleDirectoryLs maneja las peticiones para listar contenido de directorios en formato JSON
func HandleDirectoryLs(c *gin.Context) {
	// Obtener parámetros de la consulta
	diskPath := c.Query("disk")
	partitionName := c.Query("partition")
	dirPath := c.Query("path")

	// Validar parámetros
	if diskPath == "" {
		c.JSON(http.StatusBadRequest, DirectoryLsResponse{
			Success: false,
			Message: "No se proporcionó la ruta del disco (parámetro 'disk')",
		})
		return
	}

	if partitionName == "" {
		c.JSON(http.StatusBadRequest, DirectoryLsResponse{
			Success: false,
			Message: "No se proporcionó el nombre de la partición (parámetro 'partition')",
		})
		return
	}

	// Si no se especificó una ruta, usar la raíz
	if dirPath == "" {
		dirPath = "/"
	}

	// Obtener el listado de archivos/directorios
	jsonContent, err := partition_operations.ListDirectory(diskPath, partitionName, dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DirectoryLsResponse{
			Success: false,
			Message: fmt.Sprintf("Error al listar el directorio: %v", err),
		})
		return
	}

	// Convertir el string JSON a un objeto para incluirlo en la respuesta
	var data interface{}
	if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
		c.JSON(http.StatusInternalServerError, DirectoryLsResponse{
			Success: false,
			Message: fmt.Sprintf("Error al procesar datos del directorio: %v", err),
		})
		return
	}

	// Devolver la información del directorio en formato JSON
	c.JSON(http.StatusOK, DirectoryLsResponse{
		Success: true,
		Message: "Contenido del directorio obtenido correctamente",
		Content: data,
	})
}
