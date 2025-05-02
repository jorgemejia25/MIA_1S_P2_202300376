package handlers

import (
	"net/http"

	partition_operations "disk.simulator.com/m/v2/internal/disk/operations/partitions"
	"disk.simulator.com/m/v2/utils"
	"github.com/gin-gonic/gin"
)

// FileContentRequest contiene los datos necesarios para leer un archivo
type FileContentRequest struct {
	DiskPath      string `json:"diskPath"`
	PartitionName string `json:"partitionName"`
	FilePath      string `json:"filePath"`
}

// HandleReadFile maneja la solicitud para leer el contenido de un archivo
func HandleReadFile(c *gin.Context) {
	var req FileContentRequest

	// Hacer bind del JSON al struct
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validar que todos los campos requeridos están presentes
	if req.DiskPath == "" || req.PartitionName == "" || req.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Se requieren diskPath, partitionName y filePath",
		})
		return
	}

	// Obtener los directorios padre y el nombre de archivo
	parentDirs, fileName := utils.GetParentDirectories(req.FilePath)

	// Usar la función existente ReadFileContent para leer el archivo
	content, err := partition_operations.ReadFileContent(req.DiskPath, req.PartitionName, parentDirs, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error al leer el archivo: " + err.Error(),
		})
		return
	}

	// Devolver el contenido del archivo
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": content,
		"name":    req.FilePath,
	})
}
