package main

import (
	"disk.simulator.com/m/v2/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/command", handlers.HandleCommand)
	r.POST("/login", handlers.HandleLogin)
	r.POST("/logout", handlers.HandleLogout)
	r.GET("/disks", handlers.HandleDisk)                      // Ruta para listar discos
	r.GET("/disks/partitions", handlers.HandleDiskPartitions) // Modificado para usar query param
	r.GET("/directory", handlers.HandleDirectoryLs)           // Nueva ruta para listar directorios en JSON
	r.POST("/read-file", handlers.HandleReadFile)             // Nueva ruta para leer contenido de archivos
	r.GET("/journaling", func(c *gin.Context) {               // Nueva ruta para obtener el journaling
		handlers.GetJournaling(c.Writer, c.Request)
	})

	r.Run() // Por defecto escucha en :8080
}
